package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/tarantool/go-tarantool/v2"
)

type tntRespMarshall struct {
	r *tarantool.Response
}

func (t tntRespMarshall) MarshalJSON() ([]byte, error) {
	data, ok := t.r.Data[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("cast error")
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data[0])
}

func (t tntRespMarshall) GetRequest() (*http.Request, error) {

	data, ok := t.r.Tuples()[0][0].([]any)[1].(map[any]any)
	if !ok {
		return nil, fmt.Errorf("cast error")
	}
	req := make(map[string]any)
	for k, v := range data {
		kStr, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("cast error")
		}

		req[kStr] = v
	}

	log.Print(req)

	return makeRequest(req)
}

func GetAll(conn *tarantool.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		obj, err := conn.Call("get_all_proxy", []interface{}{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "failed to get all")
			return
		}

		c.JSON(http.StatusOK, tntRespMarshall{r: obj})
	}
}

func Get(conn *tarantool.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id")[1:])
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, "invalid id")
			return
		}

		obj, err := conn.Call("get_proxy", []interface{}{id})
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, "failed to get")
			return
		}

		c.JSON(http.StatusOK, tntRespMarshall{r: obj})
	}
}

func Repeat(conn *tarantool.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id")[1:])
		if err != nil {
			c.JSON(http.StatusBadRequest, "invalid id")
			return
		}

		obj, err := conn.Call("get_proxy", []interface{}{id})
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, "failed to get request to repeat")
			return
		}
		tntResp := tntRespMarshall{r: obj}

		req, err := tntResp.GetRequest()
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, "failed to get request to repeat")
			return
		}

		repeatedResp, err := doClientRequest(req)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, "failed to send request to repeat")
			return
		}

		var b []byte
		if b, err = httputil.DumpResponse(repeatedResp, true); err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, "failed to dump response")
		}

		c.String(http.StatusOK, string(b))
	}
}
