package middleware

import (
	"bytes"
	"chillcat-server/pkg/response"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// 敏感词库
var sensitiveWords = map[string]string{
	"自杀": "self_harm", "自残": "self_harm", "不想活": "self_harm",
	"想死": "self_harm", "结束生命": "self_harm", "活不下去": "self_harm",
	"轻生": "self_harm", "割腕": "self_harm", "跳楼": "self_harm",
	"杀人": "violence", "打死": "violence", "弄死": "violence",
	"砍死": "violence", "枪": "violence", "炸弹": "violence",
	"恐怖": "violence", "绑架": "violence",
	"裸照": "porn", "性爱": "porn", "约炮": "porn",
	"一夜情": "porn", "做爱": "porn", "嫖": "porn", "妓女": "porn",
	"颠覆": "politics", "推翻": "politics", "暴动": "politics",
	"分裂": "politics", "毒品": "illegal", "贩毒": "illegal",
	"冰毒": "illegal",
}

// ContentModeration 内容审核中间件：检查请求 body 中是否包含敏感词
func ContentModeration() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			c.Next()
			return
		}

		content, ok := body["content"].(string)
		if !ok || content == "" {
			c.Next()
			return
		}

		for word, category := range sensitiveWords {
			if strings.Contains(content, word) {
				if category == "self_harm" {
					response.Error(c, response.ErrSelfHarm)
				} else {
					response.Error(c, response.ErrContentBlocked)
				}
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
