package util

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/blackss2/utility/convert"
	"github.com/labstack/echo"
)

func PathToString(c echo.Context, name string) (string, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return "", errors.New("require param missing : " + name)
	}
	value, err := url.QueryUnescape(param)
	if err != nil {
		return "", err
	}
	return value, nil
}

func ParamToString(c echo.Context, name string, require bool) (string, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return "", errors.New("require param missing : " + name)
		} else {
			return "", nil
		}
	}
	value, err := url.QueryUnescape(param)
	if err != nil {
		return "", err
	}
	return value, nil
}

func BodyToString(body map[string]interface{}, name string, require bool) (string, error) {
	if param, has := body[name]; has {
		if value, is := param.(string); is {
			return value, nil
		} else {
			return "", errors.New("type error : " + name)
		}
	} else {
		if require {
			return "", errors.New("require param missing : " + name)
		} else {
			return "", nil
		}
	}
}

func PathToStringList(c echo.Context, name string) ([]string, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return nil, errors.New("require param missing : " + name)
	}
	value, err := url.QueryUnescape(param)
	if err != nil {
		return nil, err
	}
	list := strings.Split(value, ",")
	return list, nil
}

func ParamToStringList(c echo.Context, name string, require bool) ([]string, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return nil, errors.New("require param missing : " + name)
		} else {
			return nil, nil
		}
	}
	value, err := url.QueryUnescape(param)
	if err != nil {
		return nil, err
	}
	list := strings.Split(value, ",")
	return list, nil
}

func BodyToStringList(body map[string]interface{}, name string, require bool) ([]string, error) {
	if param, has := body[name]; has {
		if list, is := param.([]interface{}); is {
			strList := make([]string, 0)
			for _, v := range list {
				if str, is := v.(string); is {
					strList = append(strList, str)
				} else {
					return nil, errors.New("type error : " + name)
				}
			}
			return strList, nil
		} else {
			return nil, errors.New("type error : " + name)
		}
	} else {
		if require {
			return nil, errors.New("require param missing : " + name)
		} else {
			return nil, nil
		}
	}
}

func PathToInt(c echo.Context, name string) (int64, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return 0, errors.New("require param missing : " + name)
	}

	comp := int64(0x12345678)
	value := convert.IntWith(param, comp)
	if value == comp {
		return 0, errors.New("type error : " + name)
	}
	return value, nil
}

func ParamToInt(c echo.Context, name string, require bool) (int64, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return 0, errors.New("require param missing : " + name)
		} else {
			return 0, nil
		}
	}

	comp := int64(0x12345678)
	value := convert.IntWith(param, comp)
	if value == comp {
		return 0, errors.New("type error : " + name)
	}
	return value, nil
}

func BodyToInt(body map[string]interface{}, name string, require bool) (int64, error) {
	if param, has := body[name]; has {
		comp := int64(0x12345678)
		value := convert.IntWith(param, comp)
		if value == comp {
			return 0, errors.New("type error : " + name)
		}
		return value, nil
	} else {
		if require {
			return 0, errors.New("require param missing : " + name)
		} else {
			return 0, nil
		}
	}
}

func PathToFloat(c echo.Context, name string) (float64, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return 0, errors.New("require param missing : " + name)
	}

	comp := float64(0x12345678)
	value := convert.FloatWith(param, comp)
	if value == comp {
		return 0, errors.New("type error : " + name)
	}
	return value, nil
}

func ParamToFloat(c echo.Context, name string, require bool) (float64, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return 0, errors.New("require param missing : " + name)
		} else {
			return 0, nil
		}
	}

	comp := float64(0x12345678)
	value := convert.FloatWith(param, comp)
	if value == comp {
		return 0, errors.New("type error : " + name)
	}
	return value, nil
}

func BodyToFloat(body map[string]interface{}, name string, require bool) (float64, error) {
	if param, has := body[name]; has {
		comp := float64(0x12345678)
		value := convert.FloatWith(param, comp)
		if value == comp {
			return 0, errors.New("type error : " + name)
		}
		return value, nil
	} else {
		if require {
			return 0, errors.New("require param missing : " + name)
		} else {
			return 0, nil
		}
	}
}

func PathToBool(c echo.Context, name string) (bool, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return false, errors.New("require param missing : " + name)
	}

	param = strings.ToLower(param)
	switch param {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("type error : " + name)
	}
}

func ParamToBool(c echo.Context, name string, require bool) (bool, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return false, errors.New("require param missing : " + name)
		} else {
			return false, nil
		}
	}

	param = strings.ToLower(param)
	switch param {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("type error : " + name)
	}
}

func BodyToBool(body map[string]interface{}, name string, require bool) (bool, error) {
	if param, has := body[name]; has {
		if value, is := param.(bool); is {
			return value, nil
		} else {
			return false, errors.New("type error : " + name)
		}
	} else {
		if require {
			return false, errors.New("require param missing : " + name)
		} else {
			return false, nil
		}
	}
}

func PathToTime(c echo.Context, name string) (NullTime, error) {
	param := c.Param(name)
	if len(param) == 0 {
		return NullTime{}, errors.New("require param missing : " + name)
	}

	value := convert.Time(param)
	if value == nil {
		return NullTime{}, errors.New("type error : " + name)
	}
	if value.Unix() == (time.Time{}).Unix() {
		return NullTime{}, nil
	}
	return NullTime{Time: (*value), Valid: true}, nil
}

func ParamToTime(c echo.Context, name string, require bool) (NullTime, error) {
	param := c.QueryParam(name)
	if len(param) == 0 {
		if require {
			return NullTime{}, errors.New("require param missing : " + name)
		} else {
			return NullTime{}, nil
		}
	}

	value := convert.Time(param)
	if value == nil {
		return NullTime{}, errors.New("type error : " + name)
	}
	if value.Unix() == (time.Time{}).Unix() {
		return NullTime{}, nil
	}
	return NullTime{Time: (*value), Valid: true}, nil
}

func BodyToTime(body map[string]interface{}, name string, require bool) (NullTime, error) {
	if param, has := body[name]; has {
		value := convert.Time(param)
		if value == nil {
			return NullTime{}, errors.New("type error : " + name)
		}
		if value.Unix() == (time.Time{}).Unix() {
			return NullTime{}, nil
		}
		return NullTime{Time: (*value), Valid: true}, nil
	} else {
		if require {
			return NullTime{}, errors.New("require param missing : " + name)
		} else {
			return NullTime{}, nil
		}
	}
}

func BodyToStruct(body io.ReadCloser, ptr interface{}) error {
	defer body.Close()

	err := json.NewDecoder(body).Decode(&ptr)
	if err != nil {
		return err
	}
	return nil
}
