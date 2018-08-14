#!/usr/bin/env bash

######
# setup.sh
#
# Sets up infrastructure for a continuously-built and deployed container-based
# web app for Azure App Service.
#
# $1: repo_name: Name of GitHub repo, e.g. joshgav/go-sample.
# $2: base_name: A default prefix for all Azure resources. Defaults to $AZURE_BASE_NAME.
# $3: app_name: Name of App Service web app to use or create. Defaults to
#     "{AZURE_BASE_NAME}-webapp".
# $4: plan_name: Name of App Service plan to use or create. Defaults to
#     "{AZURE_BASE_NAME}-plan".
# $5: group_name: Name of Azure resource group to use for all resources.
#     Defaults to "${AZURE_BASE_NAME-group".
# $6: location: Name of Azure location to use for all resources. Defaults to
#     $AZURE_DEFAULT_LOCATION.
# $7: gh_token: A GitHub personal access token to set up continuous
#     integration. Defaults to $GH_TOKEN.
#####

__filename=${BASH_SOURCE[0]}
__dirname=$(cd $(dirname ${__filename}) && pwd)
__root=${__dirname}
if [[ ! -f "${__root}/.env" ]]; then cp "${__root}/.env.tpl" "${__root}/.env"; fi
if [[ -f "${__root}/.env" ]]; then source "${__root}/.env"; fi
source "${__dirname}/scripts/rm_helpers.sh"  # for ensure_group

repo_name=${1:-${REPO_NAME}}
base_name=${2:-"${AZURE_BASE_NAME}"}
app_name=${3:-"${base_name}-webapp"}
plan_name=${4:-"${base_name}-plan"}
group_name=${5:-"${base_name}-group"}
location=${6:-${AZURE_DEFAULT_LOCATION}}
gh_token=${7:-${GH_TOKEN}}

scratch_runtime_id='node|8.1'
go_runtime_id='go|1'
url_suffix=azurewebsites.net
# as specified here:
# <https://github.com/projectkudu/KuduScript/blob/master/lib/templates/deploy.bash.go.template#L8>
binary_name=app

# errors
declare -i err_nogithubtoken=102
if [[ -z ${gh_token} ]]; then
    echo 'specify a GitHub personal access token in the env var `GH_TOKEN`' \
         'to set up continuous integration'
    exit $err_nogithubtoken
fi

####

## ensure groups
ensure_group $group_name

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
    --output tsv --query id 2> /dev/null)

# Go build support is behind a flag, so we specify a scratch runtime name and then change it
# when out from flag, use these options in `create` command
#    --deployment-source-url "https://github.com/${repo_name}.git" \
#    --deployment-source-branch "master" \
if [[ -z $webapp_id ]]; then
webapp_id=$(az webapp create \
    --name "$app_name" \
    --plan ${plan_id} \
    --resource-group ${group_name} \
    --runtime "${scratch_runtime_id}" \
    --output tsv --query 'id')
fi

config_id=$(az webapp config set \
    --ids $webapp_id \
    --linux-fx-version "${go_runtime_id}" \
    --output tsv --query 'id')

# see note above, remove when Go support is out from behind flag
source_config_id=$(az webapp deployment source config \
    --ids $webapp_id \
    --repo-url "https://github.com/${repo_name}" \
    --branch 'master' \
    --git-token ${gh_token} \
    --repository-type github \
    --output tsv --query 'id')
echo "ensured web app: $webapp_id"

## ensure operation
curl -L "https://${app_name}.${url_suffix}/?name=gopherman"
echo ""
