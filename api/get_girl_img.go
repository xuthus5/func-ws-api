package handler

import (
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
	"xorm.io/xorm"
)

var engine *xorm.Engine

// MzituImg github返回的数据 只取两项 id+name
type MzituImg struct {
	Id           int64
	Catalog      string `json:"cat" xorm:"varchar(16) 'cat' comment('分类')"`
	Album        int64  `json:"album" xorm:"comment('专辑id')"`
	RawImg       string `json:"raw_img" xorm:"varchar(256) 'raw_path' comment('妹子图原始路径')"`
	BackBlazeImg string `json:"img" xorm:"varchar(256) 'backblaze_path' comment('妹子图backblaze路径')"`
}

// GirlImgResponse 是交付层的基本回应
type GirlImgResponse struct {
	Code    int         `json:"code"`    //请求状态代码
	Message string      `json:"message"` //请求结果提示
	Data    interface{} `json:"data"`    //请求结果与错误原因
}

// GirlImgWrite 输出返回结果
func GirlImgWrite(w http.ResponseWriter, response []byte) {
	//公共的响应头设置
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, OPTIONS")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(string(response))))
	_, _ = w.Write(response)
	return
}

// GirlImgConnect 数据库连接
func GirlImgConnect() error {
	var err error
	engine, err = xorm.NewEngine("sqlite3", "data/data.db")
	if err != nil {
		return err
	}
	return nil
}

// GetGirlImg 入口函数
func GetGirlImg(w http.ResponseWriter, r *http.Request) {
	if engine == nil {
		if err := GirlImgConnect(); err != nil {
			ret, _ := json.Marshal(&GirlImgResponse{
				Code:    -1,
				Message: "connect db error",
				Data:    err,
			})
			GirlImgWrite(w, ret)
			return
		}
	}
	//获取format
	var format = r.URL.Query().Get("format")

	info, err := getRandomPic()
	if err != nil {
		ret, _ := json.Marshal(&GirlImgResponse{
			Code:    500,
			Message: "query db error",
			Data:    err,
		})
		GirlImgWrite(w, ret)
		return
	}

	switch format {
	case "json":
		ret, _ := json.Marshal(&GirlImgResponse{
			Code:    200,
			Message: "ok",
			Data:    info,
		})
		GirlImgWrite(w, ret)
		return
	case "":
		w.WriteHeader(http.StatusFound)
		w.Header().Set("Location", info.BackBlazeImg)
		_, _ = w.Write(nil)
	default:
		ret, _ := json.Marshal(&GirlImgResponse{
			Code:    500,
			Message: "format arg error",
		})
		GirlImgWrite(w, ret)
		return
	}
}

// getRandomPic 获取随机一条
func getRandomPic() (*MzituImg, error) {
	var img = new(MzituImg)
	_, err := engine.SQL(`SELECT * FROM mzitu_img WHERE id >= (SELECT (select ('0.'||col) result from(select abs(random()) col)) * (SELECT MAX(id) FROM mzitu_img)) ORDER BY id LIMIT 1`).Get(img)
	if err != nil {
		return nil, err
	}

	return img, nil
}
