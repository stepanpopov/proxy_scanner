box.cfg {
    listen = '0.0.0.0:3301',
    --[[replication = {
        '127.0.0.1:3301',
        '127.0.0.1:3302',
    },
    memtx_dir  = "storage",
    wal_dir    = "storage",
    replication_connect_quorum = 1,
    replicaset_uuid = 'aaaaaaaa-0000-4000-b001-000000000000',
    instance_uuid = 'aaaaaaaa-0000-4000-a000-000000000011' ]]--
    -- checkpoint_interval = 2,
    -- checkpoint_count = 1,
}

box.schema.user.grant('guest', 'super', nil, nil, {if_not_exists=true}) 

box.schema.space.create('proxy', { if_not_exists=true })
box.space.proxy:format({
    { name='id', type='integer'},
    { name='request', type='any'}, 
    { name='response', type='any'}, 
})


box.schema.sequence.create('proxy_id')

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
