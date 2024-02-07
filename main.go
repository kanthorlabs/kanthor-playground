package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	kanthorsdk "github.com/kanthorlabs/kanthor-sdk-go"
	"github.com/kanthorlabs/kanthor-sdk-go/routing"
	"github.com/tidwall/buntdb"
)

var MessageType = "testing.playground"
var clients = map[string]*kanthorsdk.Kanthor{}
var db *buntdb.DB

func init() {
	storage := os.Getenv("STORAGE_PATH")
	if storage == "" {
		storage = "./"
	}
	year, week := time.Now().UTC().ISOWeek()
	// Open the data.db file. It will be created if it doesn't exist.
	instance, err := buntdb.Open(fmt.Sprintf("%s/playground.%d%02d.db", storage, year, week))
	if err != nil {
		panic(err)
	}
	db = instance

	credentials, err := List(db, "credentials/*")
	if err != nil {
		panic(err)
	}

	for i := range credentials {
		var item Credentials
		if err := json.Unmarshal([]byte(credentials[i]), &item); err != nil {
			log.Printf("init.sdk %v", err)
			continue
		}
		sdk, err := initSdk(item.User, item.Password)
		if err != nil {
			log.Printf("init.sdk %v", err)
			continue
		}
		clients[item.AppId] = sdk
	}
}

func main() {
	defer db.Close()
	factory := useWsFactory()

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("./views/*.html")

	router.GET("/readiness", func(c *gin.Context) {
		c.String(http.StatusOK, "ready")
	})
	router.GET("/liveness", func(c *gin.Context) {
		c.String(http.StatusOK, "live")
	})

	router.GET("/", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		wsc, err := factory()
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": fmt.Sprintf("Could not connect to the server %v", err),
				"Stack": string(debug.Stack()),
			})
			return
		}

		sdk, err := initSdk(wsc.User, wsc.Password)
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": fmt.Sprintf("Could not init the SDK %v", err),
				"Stack": string(debug.Stack()),
			})
			return
		}

		app, err := sdk.Application.Create(ctx, &kanthorsdk.ApplicationCreateReq{
			Name: fmt.Sprintf("playaround at %s", time.Now().UTC().Format(time.RFC3339)),
		})
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}

		target, err := url.Parse(os.Getenv("KANTHOR_PLAYGROUND_ENDPOINT"))
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}
		target.Path = fmt.Sprintf("/app/%s", app.Id)
		ep, err := sdk.Endpoint.Create(ctx, &kanthorsdk.EndpointCreateReq{
			AppId:  app.Id,
			Method: http.MethodPost,
			Name:   fmt.Sprintf("POST %s", target.String()),
			Uri:    target.String(),
		})
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}

		_, err = sdk.EndpointRule.Create(ctx, &kanthorsdk.EndpointRuleCreateReq{
			ConditionExpression: routing.MatchEqual(app.Id),
			ConditionSource:     routing.ConditionSourceAppId,
			EpId:                ep.Id,
			Exclusionary:        false,
			Name:                fmt.Sprintf("passthorugh all messages from the app_id:%s", app.Id),
			Priority:            100,
		})
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}
		err = Set(db, KeyEp(ep.AppId), ep)
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}

		clients[app.Id] = sdk
		// store permanently for reload
		err = Set(db, KeyWsc(ep.AppId), &Credentials{AppId: app.Id, User: wsc.User, Password: wsc.Password})
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}

		_, err = sdk.Message.Create(ctx, &kanthorsdk.MessageCreateReq{
			AppId:   app.Id,
			Body:    map[string]any{"ping": +time.Now().UnixMilli()},
			Headers: map[string]string{"X-Powered-By": "Kanthor SDK"},
			Type:    MessageType,
		})
		if err != nil {
			c.HTML(http.StatusOK, "5xx.html", gin.H{
				"Error": err.Error(),
				"Stack": string(debug.Stack()),
			})
			return
		}

		c.Redirect(http.StatusFound, target.Path)
	})

	router.GET("/app/:id", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		appId := c.Param("id")
		sdk, ok := clients[appId]
		if !ok {
			c.HTML(http.StatusOK, "4xx.html", gin.H{
				"Error": "App does not have any associated SDK",
			})
			return
		}

		app, err := sdk.Application.Get(ctx, appId)
		if err != nil {
			c.HTML(http.StatusOK, "4xx.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		messages, err := GetMessages(db, appId)
		if err != nil {
			c.HTML(http.StatusOK, "4xx.html", gin.H{
				"Error": err.Error(),
			})
			return
		}
		sdkEndpoint := os.Getenv("KANTHOR_SDK_ENDPOINT_PUBLIC")
		if sdkEndpoint == "" {
			sdkEndpoint = fmt.Sprintf("%s://%s/api", sdk.Configuration.Scheme, sdk.Configuration.Host)
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Sdk":         sdk,
			"SdkEndpoint": sdkEndpoint,
			"App":         app,
			"Auth": gin.H{
				"Authorization": sdk.Configuration.DefaultHeader["Authorization"],
			},
			"Timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
			"Messages":     messages,
			"MessageCount": len(messages),
			"MessageType":  MessageType,
		})
	})

	router.GET("/app/:id/message", func(c *gin.Context) {
		appId := c.Param("id")
		messages, err := GetMessages(db, appId)
		if err != nil {
			c.HTML(http.StatusOK, "4xx.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "message.html", gin.H{
			"Messages":     messages,
			"MessageCount": len(messages),
		})
	})

	router.GET("/app/:id/message/count", func(c *gin.Context) {
		appId := c.Param("id")
		messages, err := GetMessages(db, appId)
		if err != nil {
			c.HTML(http.StatusOK, "4xx.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "message-count.html", gin.H{
			"MessageCount": len(messages),
		})
	})

	router.POST("/app/:id", func(c *gin.Context) {
		appId := c.Param("id")
		ep, err := Get[kanthorsdk.EndpointCreateRes](db, KeyEp(appId))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		wh, err := kanthorsdk.NewWebhook(ep.SecretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		msgId := c.Request.Header.Get(kanthorsdk.HeaderWebhookId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "message-id header must be set"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := wh.Verify(body, c.Request.Header); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data := gin.H{
			"id":        c.Request.Header.Get("Webhook-Id"),
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
			"headers":   stringify(c.Request.Header),
			"body":      string(body),
		}
		err = SetStringExpire(db, KeyMsg(appId, msgId), stringify(data), time.Hour*24)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg_id": msgId, "timestamp": time.Now().UTC().Format(time.RFC3339Nano)})
	})

	router.GET("/printout", UsePrintout())
	router.POST("/printout", UsePrintout())
	router.PATCH("/printout", UsePrintout())
	router.PUT("/printout", UsePrintout())

	router.Run(":9081")
}

func UsePrintout() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		headers := c.Request.Header
		body, _ := io.ReadAll(c.Request.Body)

		isWrite := method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut
		if isWrite {
			id := fmt.Sprintf("%d", time.Now().UnixNano())
			data := map[string]any{"method": method, "headers": headers, "body": string(body), "timestamp": time.Now().Format(time.RFC3339)}
			_ = SetStringExpire(db, KeyPrintout(id), stringify(data), time.Hour)
			c.Status(http.StatusCreated)
			return
		}

		items, _ := GetPrintoutItems(db)
		c.HTML(http.StatusOK, "printout.html", gin.H{
			"ItemCount": len(items),
			"Items":     items,
		})
	}
}

func KeyPrintout(id string) string {
	return fmt.Sprintf("printout/%s", id)
}

func KeyMsg(appId, id string) string {
	return fmt.Sprintf("%s/message/%s", appId, id)
}

func KeyEp(appId string) string {
	return fmt.Sprintf("%s/ep", appId)
}

func KeyWsc(appId string) string {
	return fmt.Sprintf("credentials/%s/wsc", appId)
}

func Set(db *buntdb.DB, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return SetString(db, key, string(data))
}

func SetString(db *buntdb.DB, key string, value string) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, nil)
		return err
	})
}

func SetStringExpire(db *buntdb.DB, key string, value string, ttl time.Duration) error {
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, &buntdb.SetOptions{Expires: true, TTL: ttl})
		return err
	})
}

func List(db *buntdb.DB, key string) ([]string, error) {
	items := make([]string, 0)
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.DescendKeys(key, func(k, v string) bool {
			items = append(items, v)
			return true
		})
	})
	if errors.Is(err, buntdb.ErrNotFound) {
		return items, nil
	}
	return items, err
}

func Get[T any](db *buntdb.DB, key string) (*T, error) {
	var entity T

	err := db.View(func(tx *buntdb.Tx) error {
		data, err := tx.Get(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(data), &entity)
	})

	return &entity, err
}

func stringify(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func GetMessages(db *buntdb.DB, appId string) ([]any, error) {
	messages, err := List(db, KeyMsg(appId, "*"))
	if err != nil {
		return nil, err
	}
	items := []any{}
	for i := range messages {
		var item any
		if err := json.Unmarshal([]byte(messages[i]), &item); err != nil {
			log.Printf("unmarshal message %v", err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

type Credentials struct {
	AppId    string `json:"app_id"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type WorkspaceCreateRes struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`

	OwnerId string `json:"owner_id"`
	Name    string `json:"name"`
	Tier    string `json:"tier"`
}

type WorkspaceCredentialsCreateRes struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func initSdk(user, pass string) (*kanthorsdk.Kanthor, error) {
	options := []kanthorsdk.Option{}
	if host := os.Getenv("KANTHOR_SDK_HOST"); host != "" {
		options = append(options, kanthorsdk.WithHost(host))
	}
	return kanthorsdk.New(fmt.Sprintf("%s:%s", user, pass), options...)
}

func useWsFactory() func() (*WorkspaceCredentialsCreateRes, error) {
	client := resty.New().
		SetTimeout(time.Millisecond * 15000).
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() >= http.StatusInternalServerError
		})
	uri := os.Getenv(("KANTHOR_PORTAL_ENDPOINT"))
	token := os.Getenv(("KANTHOR_PORTAL_AUTH_CREDENTIALS"))
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic %s", token)},
	}

	return func() (*WorkspaceCredentialsCreateRes, error) {
		wsres, err := client.R().
			SetHeaderMultiValues(headers).
			SetHeader("idempotency-key", uuid.NewString()).
			SetBody(map[string]any{
				"name": fmt.Sprintf("playground at %s", time.Now().UTC().Format(time.RFC3339)),
			}).
			Post(uri + "/workspace")

		if err != nil {
			return nil, err
		}

		var ws WorkspaceCreateRes
		if err := json.Unmarshal(wsres.Body(), &ws); err != nil {
			return nil, err
		}

		wscres, err := client.R().
			SetHeaderMultiValues(headers).
			SetHeader("idempotency-key", uuid.NewString()).
			SetHeader("x-authorization-workspace", ws.Id).
			SetBody(map[string]any{
				"name":       fmt.Sprintf("playground at %s", time.Now().UTC().Format(time.RFC3339)),
				"expired_at": time.Now().UTC().Add(time.Hour * 24 * 365).UnixMilli(),
			}).
			Post(uri + "/credentials")
		if err != nil {
			return nil, err
		}

		var wsc WorkspaceCredentialsCreateRes
		if err := json.Unmarshal(wscres.Body(), &wsc); err != nil {
			return nil, err
		}

		return &wsc, nil
	}
}

func GetPrintoutItems(db *buntdb.DB) ([]any, error) {
	records, err := List(db, KeyPrintout("*"))
	if err != nil {
		return nil, err
	}
	items := []any{}
	for i := range records {
		var item any
		if err := json.Unmarshal([]byte(records[i]), &item); err != nil {
			log.Printf("unmarshal printout record %v", err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}
