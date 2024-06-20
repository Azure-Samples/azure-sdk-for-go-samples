#!/usr/bin/env bash

######
# setup.sh
#
# Sets up infrastructure for a continuously-built and deployed container-based
# web app for Azure App Service.
#
# $1: image_name: Name of image (aka repo) to use or create. Defaults to
#     $IMAGE_NAME.
# $2: image_tag: Name of image tag. Defaults to $IMAGE_TAG
# $3: repo_url: Source code repo to use for continuous build. Defaults to
#     "https://github.com/{IMAGE_NAME}.git"
# $4: base_name: A default prefix for all Azure resources. Defaults to $AZURE_BASE_NAME.
# $5: registry_name: Name of Azure container registry to use or create.
#     Defaults to "${AZURE_BASE_NAME}-registry".
# $6: app_name: Name of App Service web app to use or create. Defaults to
#     "{AZURE_BASE_NAME}-webapp".
# $7: plan_name: Name of App Service plan to use or create. Defaults to
#     "{AZURE_BASE_NAME}-plan".
# $8: group_name: Name of Azure resource group to use for all resources.
#     Defaults to "${AZURE_BASE_NAME-group".
# $9: location: Name of Azure location to use for all resources. Defaults to
#     $AZURE_DEFAULT_LOCATION.
#####

__filename=${BASH_SOURCE[0]}
__dirname=$(cd $(dirname ${__filename}) && pwd)
__root=${__dirname}
if [[ ! -f "${__root}/.env" ]]; then cp "${__root}/.env.tpl" "${__root}/.env"; fi
if [[ -f "${__root}/.env" ]]; then source "${__root}/.env"; fi
source "${__dirname}/scripts/rm_helpers.sh"  # for ensure_group

image_name=${1:-${IMAGE_NAME}}
image_tag=${2:-${IMAGE_TAG}}
repo_url=${3:-"https://github.com/${image_name}.git"}

base_name=${4:-"${AZURE_BASE_NAME}"}
registry_name=${5:-"$(echo ${base_name} | sed 's/[- _]//g')registry"}
app_name=${6:-"${base_name}-webapp"}
plan_name=${7:-"${base_name}-plan"}
group_name=${8:-"${base_name}-group"}
location=${9:-${AZURE_DEFAULT_LOCATION}}

# set after getting registry config
image_uri=
registry_sku=Standard
url_suffix=azurewebsites.net

# errors
declare -i err_registrynameexists=101
declare -i err_nogithubtoken=102

####

## ensure groups
ensure_group $group_name

## ensure registry
registry_id=$(az acr show \
        --name ${registry_name} --resource-group ${group_name} \
        --output tsv --query 'id' 2> /dev/null)

if [[ -z "${registry_id}" ]]; then
    namecheck_results=$(az acr check-name --name ${registry_name} \
        --output tsv --query '[nameAvailable, reason]')
    name_available=$(echo $namecheck_results | cut -d " " -f1)
    reason=$(echo $namecheck_results | cut -d " " -f2)
    if [[ "false" == "${name_available}" ]]; then
        echo "registry name [${registry_name}] unavailable, reason [${reason}]"
        exit $err_registrynameexists
    fi

    registry_id=$(az acr create \
        --name ${registry_name} \
        --resource-group ${group_name} \
        --sku ${registry_sku} \
        --admin-enabled 'true' \
        --location $location \
        --output tsv --query id)
fi
registry_prefix=$(az acr show \
    --name ${registry_name} --resource-group ${group_name} --output tsv --query 'loginServer')
registry_password=$(az acr credential show \
    --name ${registry_name} --output tsv --query 'passwords[0].value')
registry_username=$(az acr credential show \
    --name ${registry_name} --output tsv --query 'username')
image_uri=${registry_prefix}/${image_name}:${image_tag}
echo "ensured registry: ${registry_id}"
echo "using image_uri: ${image_uri}"


## ensure App Service plan
plan_id=$(az appservice plan show \
    --name ${plan_name} \
    --resource-group ${group_name} \
    --output tsv --query id)

if [[ -z $plan_id ]]; then
    plan_id=$(az appservice plan create \
        --name ${plan_name} \
        --resource-group ${group_name} \
        --location $location \
        --is-linux \
        --output tsv --query id)
fi
echo "ensured plan $plan_id"

## ensure Web App
webapp_id=$(az webapp show \
    --name ${app_name} \
    --resource-group ${group_name} \
    --output tsv --query id)

if [[ -z $webapp_id ]]; then
webapp_id=$(az webapp create \
    --name "$app_name" \
    --plan ${plan_id} \
    --resource-group ${group_name} \
    --deployment-container-image-name ${image_uri} \
    --output tsv --query 'id')
fi

# set up web app for continuous deployment
webapp_config=$(az webapp config container set \
    --ids $webapp_id \
    --docker-custom-image-name ${image_uri} \
    --docker-registry-server-url "https://${registry_prefix}" \
    --docker-registry-server-user ${registry_username} \
    --docker-registry-server-password ${registry_password} \
    --output json)
webapp_config2=$(az webapp deployment container config \
    --ids $webapp_id \
    --enable-cd 'true' \
    --output json)

echo -e "webapp_config:\n$webapp_config"
echo -e "webapp_config2:\n$webapp_config2"
echo "ensured web app: $webapp_id"

curl -L --fail "${repo_url}" 2> /dev/null 1> /dev/null
curl_exitcode=$?

if [[ "$curl_exitcode" == "22" ]]; then
    echo "could not reach hosted repo, instead building locally and pushing"
    echo "continuous build and deploy requires a hosted repo"
    # run one build and push image
    # add `--no-logs` to suppress log output
    build_id=$(az acr build \
         --registry ${registry_name} \
         --resource-group ${group_name} \
         --file 'Dockerfile' \
         --image "${image_name}:${image_tag}" \
         --os 'Linux' \
         --output tsv --query id \
         ${__root})

else
    echo "using hosted repo: $repo_url for continuous build and deploy"

    if [[ -z ${GH_TOKEN} ]]; then
        echo 'specify a GitHub personal access token in the env var `GH_TOKEN`' \
             'to set up continuous deploy'
        exit $err_nogithubtoken
    fi

    # set up a build task to build on commit
    buildtask_name=buildoncommit
    buildtask_id=$(az acr build-task create \
        --name ${buildtask_name} \
        --registry ${registry_name} \
        --resource-group ${group_name} \
        --context ${repo_url} \
        --git-access-token ${GH_TOKEN} \
        --image "${image_name}:${image_tag}" \
        --branch 'master' \
        --commit-trigger-enabled 'true' \
        --file 'Dockerfile' \
        --os 'Linux' \
        --output tsv --query id)

    # and run once now
    # add `--no-logs` to suppress log output
    buildtask_run_id=$(az acr build-task run \
        --name ${buildtask_name} \
        --registry ${registry_name} \
        --resource-group ${group_name} \
        --output tsv --query id)
fi

## ensure operation
curl -L "https://${app_name}.${url_suffix}/?name=gopherman"
echo ""
