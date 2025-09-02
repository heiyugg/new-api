package controller

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/model"
	"one-api/setting/system_setting"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type NodelocOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type NodelocUser struct {
	Sub               string `json:"sub"`
	Username          string `json:"preferred_username"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Picture           string `json:"picture"`
	Groups            []string `json:"groups"`
}

func getNodelocUserInfoByCode(code string, c *gin.Context) (*NodelocUser, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}

	settings := system_setting.GetNodelocOAuthSettings()
	if !settings.Enabled {
		return nil, errors.New("nodeloc oauth is disabled")
	}

	// Get redirect URI
	redirectURI := settings.RedirectUri
	if redirectURI == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		redirectURI = fmt.Sprintf("%s://%s/oauth/nodeloc", scheme, c.Request.Host)
	}

	// Get access token using Basic auth
	credentials := settings.ClientId + ":" + settings.ClientSecret
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", settings.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to connect to Nodeloc server")
	}
	defer res.Body.Close()

	var tokenRes NodelocOAuthResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	if tokenRes.AccessToken == "" {
		return nil, fmt.Errorf("failed to get access token")
	}

	// Get user info
	userEndpoint := settings.UserInfoEndpoint
	req, err = http.NewRequest("GET", userEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
	req.Header.Set("Accept", "application/json")

	res2, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to get user info from Nodeloc")
	}
	defer res2.Body.Close()

	var nodelocUser NodelocUser
	if err := json.NewDecoder(res2.Body).Decode(&nodelocUser); err != nil {
		return nil, err
	}

	if nodelocUser.Sub == "" {
		return nil, errors.New("invalid user info returned")
	}

	return &nodelocUser, nil
}

func NodelocOAuth(c *gin.Context) {
	session := sessions.Default(c)

	errorCode := c.Query("error")
	if errorCode != "" {
		errorDescription := c.Query("error_description")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('OAuth错误: %s');
					window.close();
				</script>
			</body>
			</html>
		`, errorDescription)))
		return
	}

	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('状态验证失败，请重试');
					window.close();
				</script>
			</body>
			</html>
		`))
		return
	}

	username := session.Get("username")
	if username != nil {
		NodelocBind(c)
		return
	}

	if !system_setting.GetNodelocOAuthSettings().Enabled {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('管理员未开启通过 Nodeloc 登录以及注册');
					window.close();
				</script>
			</body>
			</html>
		`))
		return
	}

	code := c.Query("code")
	nodelocUser, err := getNodelocUserInfoByCode(code, c)
	if err != nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('获取用户信息失败: %s');
					window.close();
				</script>
			</body>
			</html>
		`, err.Error())))
		return
	}

	user := model.User{
		NodelocId: nodelocUser.Sub,
	}

	// Check if user exists
	if model.IsNodelocIdAlreadyTaken(user.NodelocId) {
		err := user.FillUserByNodelocId()
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>登录失败</title>
					<meta charset="utf-8">
				</head>
				<body>
					<script>
						alert('获取用户信息失败: %s');
						window.close();
					</script>
				</body>
				</html>
			`, err.Error())))
			return
		}
		if user.Id == 0 {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>登录失败</title>
					<meta charset="utf-8">
				</head>
				<body>
					<script>
						alert('用户已注销');
						window.close();
					</script>
				</body>
				</html>
			`))
			return
		}
	} else {
		if common.RegisterEnabled {
			if nodelocUser.Username != "" {
				user.Username = "nodeloc_" + nodelocUser.Username
			} else {
				user.Username = "nodeloc_" + strconv.Itoa(model.GetMaxUserId()+1)
			}
			if nodelocUser.Name != "" {
				user.DisplayName = nodelocUser.Name
			} else {
				user.DisplayName = "Nodeloc User"
			}
			user.Email = nodelocUser.Email
			user.Role = common.RoleCommonUser
			user.Status = common.UserStatusEnabled

			affCode := session.Get("aff")
			inviterId := 0
			if affCode != nil {
				inviterId, _ = model.GetUserIdByAffCode(affCode.(string))
			}

			if err := user.Insert(inviterId); err != nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
					<!DOCTYPE html>
					<html>
					<head>
						<title>登录失败</title>
						<meta charset="utf-8">
					</head>
					<body>
						<script>
							alert('用户注册失败: %s');
							window.close();
						</script>
					</body>
					</html>
				`, err.Error())))
				return
			}
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>登录失败</title>
					<meta charset="utf-8">
				</head>
				<body>
					<script>
						alert('管理员关闭了新用户注册');
						window.close();
					</script>
				</body>
				</html>
			`))
			return
		}
	}

	if user.Status != common.UserStatusEnabled {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('用户已被封禁');
					window.close();
				</script>
			</body>
			</html>
		`))
		return
	}

	// 设置登录会话
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("group", user.Group)
	err = session.Save()
	if err != nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>登录失败</title>
				<meta charset="utf-8">
			</head>
			<body>
				<script>
					alert('无法保存会话信息，请重试');
					window.close();
				</script>
			</body>
			</html>
		`))
		return
	}

	// 成功登录后关闭窗口并刷新父窗口
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>登录成功</title>
			<meta charset="utf-8">
		</head>
		<body>
			<script>
				alert('登录成功！');
				if (window.opener) {
					window.opener.location.reload();
				}
				window.close();
			</script>
		</body>
		</html>
	`))
}

func NodelocBind(c *gin.Context) {
	if !system_setting.GetNodelocOAuthSettings().Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Nodeloc 登录以及注册",
		})
		return
	}

	code := c.Query("code")
	nodelocUser, err := getNodelocUserInfoByCode(code, c)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	user := model.User{
		NodelocId: nodelocUser.Sub,
	}

	if model.IsNodelocIdAlreadyTaken(user.NodelocId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Nodeloc 账户已被绑定",
		})
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)

	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	user.NodelocId = nodelocUser.Sub
	err = user.Update(false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
}
