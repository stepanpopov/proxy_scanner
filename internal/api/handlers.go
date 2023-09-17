package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/tarantool/go-tarantool/v2"
)

type tntRespMarshall struct {
	r *tarantool.Response
}

func (t tntRespMarshall) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(t.r.Data)
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
			c.JSON(http.StatusInternalServerError, "failed to get")
			return
		}

		c.JSON(http.StatusOK, tntRespMarshall{r: obj})
	}
}

/* func Repeat(conn *tarantool.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, "invalid id")
			return
		}

		obj, err := conn.Call("get_proxy", []interface{}{id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "failed to get")
			return
		}

		c.JSON(http.StatusOK, obj)
	}
} */
