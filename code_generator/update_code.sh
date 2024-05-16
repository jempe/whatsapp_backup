#!/bin/bash
#
set -e # exit on error

BASEDIR=..
APITEMPLATESDIR=$GOPATH/src/github.com/jempe/api_template/templates
GENERATOR=api_code_generator

SEDBINARY=sed

ISMAC=$(uname -a | grep Darwin)

if [ -n "$ISMAC" ]; then
	SEDBINARY=gsed
fi

if [ ! -f $GOBIN/$GENERATOR ]; then
    echo "building generator"
    go install github.com/jempe/api_template/cmd/api_code_generator
    exit 1
fi

FOLDERLIST=(
	cmd/api
	internal/data
	internal/jsonlog
	internal/validator
)
for FOLDER in "${FOLDERLIST[@]}"
do
	if [ ! -f $BASEDIR/$FOLDER ]; then
		echo "creating folder $BASEDIR/$FOLDER"
	    	mkdir -p $BASEDIR/$FOLDER
	fi
done


# list of files to generate
FILESLIST=(
	cmd/api/main.go
	cmd/api/errors.go
	cmd/api/middleware.go
	cmd/api/healthcheck.go
	cmd/api/db.go
	cmd/api/helpers.go
	cmd/api/routes.go
	cmd/api/embeddings.go
	internal/jsonlog/jsonlog.go
	internal/data/filters.go
	internal/data/models.go
	internal/data/embeddings.go
	internal/validator/validator.go
)

for FILE in "${FILESLIST[@]}"
do
	FILEFULL=$BASEDIR/$FILE
	echo "generating $FILEFULL"
	$GENERATOR -schema schema.json -table messages -overwrite  -output $FILEFULL $APITEMPLATESDIR/$FILE.tmpl
	gofmt -w $FILEFULL
done

echo ""
echo "generating tables files"
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


	echo "generating internal/data custom file of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/internal/data/$TABLE"_custom.go" $APITEMPLATESDIR/internal/data/items_custom.go.tmpl
	gofmt -w $BASEDIR/internal/data/$TABLE"_custom.go"


	echo "generating cmd/api files of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/cmd/api/$TABLE.go $APITEMPLATESDIR/cmd/api/items.go.tmpl
	gofmt -w $BASEDIR/cmd/api/$TABLE.go

	echo "generating cmd/api custom files of $TABLE"
	$GENERATOR -schema schema.json -table $TABLE -overwrite -output $BASEDIR/cmd/api/$TABLE"_custom.go" $APITEMPLATESDIR/cmd/api/items_custom.go.tmpl
	gofmt -w $BASEDIR/cmd/api/$TABLE.go
done

#custom routes start

#ROUTESFILE=$BASEDIR/cmd/api/routes.go

#$SEDBINARY -i '/\/\/custom_routes/ {
#	r cmd/api/routes_custom.go.tmpl
#	d
#}' $ROUTESFILE

#gofmt -w $ROUTESFILE

#custom routes end

#custom code

#ABTESTSCUSTOMFILE=$BASEDIR/cmd/api/abtests_custom.go

#$SEDBINARY -i '/\/\/custom_code/ {
#	r cmd/api/abtests_custom.go.tmpl
#	d
#}' $ABTESTSCUSTOMFILE

#gofmt -w $ABTESTSCUSTOMFILE

#custom code end

#copy scripts start

cp -r $APITEMPLATESDIR/../scripts $BASEDIR/

#copy scripts end

# generate semantic search table

FILEFULL=$BASEDIR/cmd/api/cronjob.go

echo "generating semantic search table files for $SEMANTICSEARCHTABLE"
$GENERATOR -schema schema.json -table $SEMANTICSEARCHTABLE -overwrite  -output $FILEFULL $APITEMPLATESDIR/cmd/api/cronjob.go.tmpl

SEMANTICFILESLIST=(
cmd/api/cronjob.go
cmd/api/phrases_custom.go
internal/data/phrases_custom.go
)

for FILE in "${SEMANTICFILESLIST[@]}"
do
	FILEFULL=$BASEDIR/$FILE
	echo "replacing provider  $FILEFULL"
	$SEDBINARY -i 's/phrase__provider__/message/g' $FILEFULL
	$SEDBINARY -i 's/phrases__provider__/messages/g' $FILEFULL
	$SEDBINARY -i 's/Phrase__provider__/Message/g' $FILEFULL
	$SEDBINARY -i 's/Phrases__provider__/Messages/g' $FILEFULL
	$SEDBINARY -i 's/--CUSTOM_JOIN_PLACEHOLDER/\t\t--ADD CUSTOM JOINS HERE/g' $FILEFULL
	gofmt -w $FILEFULL
done

echo "Files generated successfully"

