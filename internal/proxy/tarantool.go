package proxy

import "github.com/tarantool/go-tarantool/v2"

type TarantoolProxy struct {
	conn *tarantool.Connection
}

func NewTarantoolProxy(conn *tarantool.Connection) *TarantoolProxy {
	return &TarantoolProxy{conn: conn}
}

func (t TarantoolProxy) insertReqResp(req, resp string) error {
	_, err := t.conn.Call("insert_proxy", []interface{}{req, resp})
	return err
}
