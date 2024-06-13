function ensure_group () {
    local group_name=$1
    local group_id

    group_id=$(az group show --name $group_name --query 'id' --output tsv 2> /dev/null)

    if [[ -z $group_id ]]; then
        group_id=$(az group create \
            --name $group_name \
            --location $location \
            --query 'id' --output tsv)
    fi
    echo $group_id
}

