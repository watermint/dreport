#!/bin/sh

export TARGET_OS="windows darwin linux"
BUILD=`pwd`/out
DIST=/dist

if [ ! -e $PROJECT_ROOT/credentials ]; then
  echo Credential file not found. Creating dummy file for CI
  cp $PROJECT_ROOT/credentials.sample $PROJECT_ROOT/credentials
fi

APP_VERSION=$(cat $PROJECT_ROOT/version)
CREDENTIALS=$(cat $PROJECT_ROOT/credentials | xargs)

echo building: $APP_VERSION
echo UID: `id`

for t in $TARGET_OS; do
  mkdir -p "$BUILD/$t";
done

cd $PROJECT_ROOT
glide install

echo Testing...
go test  $(glide novendor)
if [ x"$?" != x"0" ]; then
  echo Test failed: $?
  exit 1
fi

for t in $TARGET_OS; do
  echo Building: $t
  cd $BUILD/$t
  GOOS=$t GOARCH=amd64 go build -ldflags "-X main.AppVersion=$APP_VERSION $CREDENTIALS" github.com/watermint/dreport
done

cd $BUILD
zip -9 -r $DIST/dreport-$APP_VERSION.zip .
