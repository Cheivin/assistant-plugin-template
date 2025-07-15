package sse

import (
	"assistant-plugin-go/hub"
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type connector struct {
	ctx    context.Context
	r      *resty.Client
	config struct {
		host     string
		port     int
		username string
		password string
	}
	connected bool
}

type result[T any] struct {
	Code int    `json:"code"` // 0表示成功
	Msg  string `json:"msg"`  //
	Data T      `json:"data"`
}

type Option func(c *connector)

func WithAuth(username, password string) Option {
	return func(c *connector) {
		c.config.username = username
		c.config.password = password
		c.r = c.r.SetBasicAuth(username, password)
	}
}

func New(ctx context.Context, host string, port int, options ...Option) hub.Connector {
	c := &connector{
		ctx: ctx,
		r:   resty.New().SetBaseURL("http://" + host + ":" + strconv.Itoa(port)).SetDisableWarn(true),
		config: struct {
			host     string
			port     int
			username string
			password string
		}{host: host, port: port},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *connector) Alive() bool {
	return c.connected
}

func (c *connector) GetGroupMembers(gid string) hub.GroupMembers {
	resp, err := c.r.R().
		SetQueryParam("gid", gid).
		SetResult(result[hub.GroupMembers]{}).
		Get("group")
	if err != nil {
		slog.Error("SSEClient 获取群成员列表失败", "err", err)
		return nil
	}
	if !resp.IsSuccess() {
		slog.Error("SSEClient 获取群成员列表失败", "response", string(resp.Body()))
		return nil
	}
	res := resp.Result().(*result[hub.GroupMembers])
	if res == nil {
		slog.Error("SSEClient 获取群成员列表失败,结果为nil", "response", string(resp.Body()))
		return nil
	}
	if res.Code != 0 {
		slog.Error("SSEClient 获取群成员列表失败", "err", err)
		return nil
	}
	return res.Data
}

func (c *connector) GetGroupMember(gid string, uid string) *hub.GroupMember {
	resp, err := c.r.R().
		SetQueryParam("gid", gid).
		SetQueryParam("uid", uid).
		SetResult(result[*hub.GroupMember]{}).
		Get("group/member")
	if err != nil {
		slog.Error("SSEClient 获取群成员信息失败", "err", err)
		return nil
	}
	if !resp.IsSuccess() {
		slog.Error("SSEClient 获取群成员信息失败", "response", string(resp.Body()))
		return nil
	}
	res := resp.Result().(*result[*hub.GroupMember])
	if res == nil {
		slog.Error("SSEClient 获取群成员信息失败,结果为nil", "response", string(resp.Body()))
		return nil
	}
	if res.Code != 0 {
		slog.Error("SSEClient 获取群成员信息失败", "err", err)
		return nil
	}
	return res.Data
}

func (c *connector) Listen(f func(hub.Message)) {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.connected = false
			slog.Info("SSEClient 开始连接监听")
			if err := c.listen(c.ctx, func(message hub.Message) {
				defer func() {
					if err := recover(); err != nil {
						slog.Error("SSEClient recover 处理消息出错", "err", err)
					}
				}()
				f(message)
			}); errors.Is(err, context.Canceled) {
				return
			} else {
				slog.Error("SSEClient 连接失败,5秒后重试", "err", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (c *connector) listen(ctx context.Context, f func(message hub.Message)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.connected = true
	resp, err := c.r.R().
		SetContext(ctx).
		SetHeader("Accept", "text/event-stream").
		SetDoNotParseResponse(true).
		Get("sse")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.Status())
	}

	defer resp.RawBody().Close()
	// 处理事件流数据
	scanner := bufio.NewScanner(resp.RawBody())

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			slog.Info("SSEClient 流退出", "err", err)
			return err
		default:
			if !scanner.Scan() {
				return scanner.Err()
			}
			line := strings.Trim(strings.TrimPrefix(scanner.Text(), "data:"), "\n")
			if line != "" {
				message := hub.Message{}
				if err := json.Unmarshal([]byte(line), &message); err != nil {
					slog.Error("SSEClient 消息反序列化失败", "line", line, "err", err)
				} else {
					f(message)
				}
			}
		}
	}
}

func (c *connector) sendMsg(typ int, gid, body, filename string) (int64, error) {
	resp, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"gid":      gid,
			"type":     typ,      // 回复类型 1:文本,2:图片,3:视频,4:文件
			"body":     body,     // 回复内容,type=1时为文本内容,type=2/3/4时为资源地址
			"filename": filename, // 文件名称
		}).
		SetResult(result[json.Number]{}).
		Post("/msg/send")
	if err != nil {
		return 0, err
	}
	if !resp.IsSuccess() {
		return 0, errors.New(resp.Status())
	}
	res := resp.Result().(*result[json.Number])
	if res == nil {
		return 0, errors.New("发送失败,结果为nil")
	}
	if res.Code != 0 {
		return 0, errors.New(res.Msg)
	}
	number, _ := res.Data.Int64()
	return number, nil
}
func (c *connector) SendText(gid, msg string) (int64, error) {
	return c.sendMsg(1, gid, msg, "")
}

func (c *connector) SendImageByUrl(gid, path, filename string) (int64, error) {
	return c.sendMsg(2, gid, path, filename)
}

func (c *connector) SendImageByBase64(gid string, data []byte, filename string) (int64, error) {
	return c.sendMsg(2, gid, "BASE64:"+base64.RawStdEncoding.EncodeToString(data), filename)
}

func (c *connector) SendVideoByUrl(gid, path, filename string) (int64, error) {
	return c.sendMsg(3, gid, path, filename)
}

func (c *connector) SendVideoByBase64(gid string, data []byte, filename string) (int64, error) {
	return c.sendMsg(3, gid, "BASE64:"+base64.RawStdEncoding.EncodeToString(data), filename)
}

func (c *connector) SendFileByUrl(gid string, path, filename string) (int64, error) {
	return c.sendMsg(4, gid, path, filename)
}

func (c *connector) SendFileByBase64(gid string, data []byte, filename string) (int64, error) {
	return c.sendMsg(4, gid, "BASE64:"+base64.RawStdEncoding.EncodeToString(data), filename)
}
