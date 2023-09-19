box.cfg {
    listen = '0.0.0.0:3301',
}

box.schema.user.grant('guest', 'super', nil, nil, {if_not_exists=true}) 

box.schema.space.create('proxy', { if_not_exists=true })
box.space.proxy:format({
    { name='id', type='integer'},
    { name='request', type='any'}, 
    { name='response', type='any'}, 
})


box.schema.sequence.create('proxy_id', { if_not_exists=true })

box.space.proxy:create_index(
    'primary',
    {
        sequence = 'proxy_id',
        parts = { 'id' },
        if_not_exists = true,
    }
)

function insert_proxy(req, resp)
    box.space.proxy:insert{nil, req, resp}
end

function get_proxy(id)
    return box.space.proxy:select{ id }
end

function get_all_proxy()
    return box.space.proxy:select{}
end

-- require('console').start() os.exit()
