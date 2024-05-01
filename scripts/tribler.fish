function delete_torrent
    set api_key (cat .env | grep TRIBLER_API_KEY | tr -d '"' | cut -d'=' -f2)
    set api_endpoint (cat .env | grep TRIBLER_API_ENDPOINT | tr -d '"' | cut -d'=' -f2)
    set infohashes (get_torrents_list | fzf -m | cut -d',' -f1 | tr -d '"')
    for infohash in $infohashes
        curl -X DELETE -s "$api_endpoint/downloads/$infohash" \
            -H "Content-type: application/json" \
            --data '{ "remove_data": true }' \
            -H "X-API-KEY: $api_key"
    end
end

# Get files list of a torrent
function get_files_list
    set api_key (cat .env | grep TRIBLER_API_KEY | tr -d '"' | cut -d'=' -f2)
    set api_endpoint (cat .env | grep TRIBLER_API_ENDPOINT | tr -d '"' | cut -d'=' -f2)
    set infohash (get_torrents_list | fzf | cut -d',' -f1 | tr -d '"')
    curl -s "$api_endpoint/downloads/$infohash/files" -H "X-API-KEY: $api_key" | jq -r '.files[] | [.index, .name, .size] | @csv'
end

# Get list of torrents with the following fields: name, status, size
function get_torrents_list
    # extract api_key and api_endpoint from .env file and strip quotes
    set api_key (cat .env | grep TRIBLER_API_KEY | tr -d '"' | cut -d'=' -f2)
    set api_endpoint (cat .env | grep TRIBLER_API_ENDPOINT | tr -d '"' | cut -d'=' -f2)

    curl "$api_endpoint/downloads" -s -H "X-API-KEY: $api_key" | jq -r '.downloads[] | [.infohash, .status, .size, .name] | @csv'
end
