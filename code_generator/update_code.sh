#!/bin/bash
#

error() {
	echo "****************************************************************************************"
	echo "****************************************************************************************"
	echo ""
	echo "ERROR: $1"
	echo ""
	echo "****************************************************************************************"
	echo "****************************************************************************************"
	echo ""
}

BASEDIR=..
APITEMPLATESDIR=$GOPATH/src/github.com/jempe/api_template/templates
GENERATOR=api_code_generator
GENERATORVERSION=7
APITEMPLATESVERSION=2024-09-24_1

SEDBINARY=sed

ISMAC=$(uname -a | grep Darwin)

if [ -n "$ISMAC" ]; then
	echo ""
	echo "Mac OS detected Using gsed"
	echo ""
	SEDBINARY=gsed
else
	echo ""
	echo "Linux detected Using sed"
	echo ""
fi

set -e # exit on error

echo ""
echo "Checking API code generator"
echo ""

if [ ! -f $GOBIN/$GENERATOR ]; then
    error "API code generator not found. Please install it using: go get github.com/jempe/api_template/cmd/api_code_generator"
    exit 1
fi

ACGVERSION=$($GENERATOR -version)

if [ "$ACGVERSION" == "API Code Generator Version: $GENERATORVERSION" ]; then
	echo "API code generator version: $ACGVERSION"
else
	error "API code generator version mismatch. Expected: $GENERATORVERSION Found: $ACGVERSION \nPlease update it using: go get -u github.com/jempe/api_template/cmd/api_code_generator"
	exit 1
fi

echo ""
echo "Checking API templates"
echo ""

if [ ! -d $APITEMPLATESDIR ]; then
    error "API templates not found. Please install them by cloning the repo github.com/jempe/api_template/templates"
    exit 1
fi

if [ ! -f $APITEMPLATESDIR/version ]; then
    error "API templates version file not found. Please update them using: go get -u github.com/jempe/api_template/templates"
    exit 1
fi

if [ $(cat $APITEMPLATESDIR/version) != $APITEMPLATESVERSION ]; then
    error "API templates version mismatch. Found: $(cat $APITEMPLATESDIR/version) Expected: $APITEMPLATESVERSION"
    echo "Please update github.com/jempe/api_template"
    exit 1
fi

echo ""
echo "Creating folders"
echo ""

FOLDERLIST=(
	cmd/api
	internal/data
	internal/jsonlog
	internal/validator
	internal/mailer/templates
	#internal/llms/llmclaude
	#internal/llms/llmopenai
	#internal/llms/llmgoogle
	ui/html/pages
	ui/html/partials
	ui/static/img
	ui/static/fonts/Roboto
)
for FOLDER in "${FOLDERLIST[@]}"
do
	if [ ! -f $BASEDIR/$FOLDER ]; then
		echo "Creating folder $BASEDIR/$FOLDER"
	    	mkdir -p $BASEDIR/$FOLDER
	fi
done

echo ""
echo "Generating API code"
echo ""

# list of files to generate
FILESLIST=(
	cmd/api/context.go
	cmd/api/db.go
	cmd/api/embeddings.go
	cmd/api/errors.go
	cmd/api/handlers.go
	cmd/api/healthcheck.go
	cmd/api/helpers.go
	cmd/api/main.go
	cmd/api/middleware.go
	cmd/api/routes.go
	cmd/api/server.go
	cmd/api/templates.go
	cmd/api/tokens.go
	cmd/api/users.go
	internal/data/embeddings.go
	internal/data/filters.go
	internal/data/models.go
	internal/data/tokens.go
	internal/data/users.go
	internal/jsonlog/jsonlog.go
	internal/mailer/mailer.go
	internal/validator/validator.go
	ui/efs.go
	ui/html/pages/dashboard.tmpl
	ui/html/partials/sidebar.tmpl
	#cmd/api/llms.go
	#internal/llms/llms.go
	#internal/llms/llmclaude/claude_api_client.go
	#internal/llms/llmopenai/openai_api_client.go
)

for FILE in "${FILESLIST[@]}"
do
	FILEFULL=$BASEDIR/$FILE
	echo "Generating $FILEFULL"
	$GENERATOR -schema schema.json -table messages -overwrite  -output $FILEFULL $APITEMPLATESDIR/$FILE.tmpl

	set +e
	ISGOFILE=$(echo $FILEFULL | grep .go)

	if [ -n "$ISGOFILE" ]; then
		echo "Formatting $FILEFULL"
		echo ""
		gofmt -w $FILEFULL
	fi
	set -e

done

echo ""
echo "Generating CRUD table files"
echo ""


TABLESLIST=(
chats
messages
phrases
)

SEMANTICSEARCHTABLE=phrases

for TABLE in "${TABLESLIST[@]}"
do
	echo "generating internal/data file of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/internal/data/$TABLE.go $APITEMPLATESDIR/internal/data/items.go.tmpl
	gofmt -w $BASEDIR/internal/data/$TABLE.go

	echo "generating internal/data validation of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/internal/data/$TABLE"_validation.go" $APITEMPLATESDIR/internal/data/items_validation.go.tmpl
	gofmt -w $BASEDIR/internal/data/$TABLE.go

	echo "generating cmd/api files of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/cmd/api/$TABLE.go $APITEMPLATESDIR/cmd/api/items.go.tmpl
	gofmt -w $BASEDIR/cmd/api/$TABLE.go

	echo "generating ui/html files of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/ui/html/pages/$TABLE.tmpl $APITEMPLATESDIR/ui/html/pages/category.tmpl.tmpl
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/ui/html/pages/"$TABLE"_item.tmpl $APITEMPLATESDIR/ui/html/pages/item.tmpl.tmpl

done

#auth files start

echo ""
echo "Adding auth HTML client files"
echo ""

CLIENTFILES=(
	ui/html/base.tmpl
	ui/html/pages/auth_pages.tmpl
	ui/html/partials/fonts.tmpl
	ui/html/partials/header.tmpl
	ui/html/partials/common.tmpl
	ui/static/img/last_page_24dp_FILL0_wght400_GRAD0_opsz24.svg
	ui/static/img/chevron_backward_24dp_FILL0_wght400_GRAD0_opsz24.svg
	ui/static/img/first_page_24dp_FILL0_wght400_GRAD0_opsz24.svg
	ui/static/img/chevron_forward_24dp_FILL0_wght400_GRAD0_opsz24.svg
	ui/static/fonts/Roboto/Roboto-Medium.ttf
	ui/static/fonts/Roboto/Roboto-Light.ttf
	ui/static/fonts/Roboto/Roboto-Regular.ttf
	ui/static/fonts/Roboto/Roboto-MediumItalic.ttf
	ui/static/fonts/Roboto/Roboto-ThinItalic.ttf
	ui/static/fonts/Roboto/Roboto-BoldItalic.ttf
	ui/static/fonts/Roboto/Roboto-LightItalic.ttf
	ui/static/fonts/Roboto/Roboto-Italic.ttf
	ui/static/fonts/Roboto/LICENSE.txt
	ui/static/fonts/Roboto/Roboto-BlackItalic.ttf
	ui/static/fonts/Roboto/Roboto-Bold.ttf
	ui/static/fonts/Roboto/Roboto-Thin.ttf
	ui/static/fonts/Roboto/Roboto-Black.ttf
	internal/mailer/templates/user_welcome.tmpl
	internal/mailer/templates/token_password_reset.tmpl
	internal/mailer/templates/token_activation.tmpl
)

for FILE in "${CLIENTFILES[@]}"
do
	FILEFULL=$BASEDIR/$FILE

	echo "Copying $APITEMPLATESDIR/$FILE to $FILEFULL"

	cp $APITEMPLATESDIR/$FILE $FILEFULL
done

#auth files end

#custom routes start

echo ""
echo "Adding custom routes"
echo ""

#ROUTESFILE=$BASEDIR/cmd/api/routes.go

#$SEDBINARY -i '/\/\/custom_routes/ {
#	r cmd/api/routes_custom.go.tmpl
#	d
#}' $ROUTESFILE

#gofmt -w $ROUTESFILE

#custom routes end

#custom main start

echo ""
echo "Updating main"
echo ""

#MAINFILE=$BASEDIR/cmd/api/main.go

#$SEDBINARY -i '/\/\/custom_config_variables/ {
#	r cmd/api/main_custom_config.go.tmpl
#	d
#}' $MAINFILE

#$SEDBINARY -i '/\/\/custom_config_flags/ {
#	r cmd/api/main_custom_config_flags.go.tmpl
#	d
#}' $MAINFILE

#gofmt -w $MAINFILE

#custom main end

#custom code

#custom code end

# generate semantic search table

FILEFULL=$BASEDIR/cmd/api/cronjob.go

echo ""
echo "Generating semantic search table files for $SEMANTICSEARCHTABLE"
echo ""

$GENERATOR -schema schema.json -table $SEMANTICSEARCHTABLE -overwrite  -output $FILEFULL $APITEMPLATESDIR/cmd/api/cronjob.go.tmpl

gofmt -w $FILEFULL

