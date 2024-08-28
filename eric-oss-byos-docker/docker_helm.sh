#!/bin/bash
IMAGE_NAME="$1"
BUILD_ID="$2"
#HELM_CHART="$3"
SIM_NAME="$3"
CHART_NAME="$4"

#helm_link="$3"

#wget "$helm_link"

#directory_name=$(basename "$helm_link")

#tar -xvf "$directory_name"

dockerfile_path="postgres/"

docker build -t "$IMAGE_NAME:$BUILD_ID" "$dockerfile_path".

if [ $? -eq 0 ]; then
echo "Docker image '$IMAGE_NAME:$BUILD_ID' was built successfully."
else
echo "Failed to build Docker image."
fi

docker push "$IMAGE_NAME:$BUILD_ID"

echo "Image name '$IMAGE_NAME:$BUILD_ID' was pushed successfully."

cd /postgres

#chart_dir="$4"

sed -i "s|datasetimage:.*|datasetimage: $IMAGE_NAME:$BUILD_ID|" "$CHART_NAME/values.yaml"

if [ $? -eq 0 ]; then
echo "Image name '$IMAGE_NAME:$BUILD_ID' replaced in the Helm chart."
else
echo "Failed to replace Image name in the Helm chart."
fi

#chart_name="$4"

sed -i "s|name:.*|name: $SIM_NAME|" "$CHART_NAME/Chart.yaml"

if [ $? -eq 0 ]; then
echo "Chart_name '$SIM_NAME' replaced in the Helm chart."
else
echo "Failed to replace Chart_name in the Helm chart."
fi

#chart_version="$4"

sed -i "s|version:.*|version: $BUILD_ID|" "$CHART_NAME/Chart.yaml"

if [ $? -eq 0 ]; then
echo "Chart_version '$version:$BUILD_ID' replaced in the Helm chart."
else
echo "Failed to replace version in the Helm chart."
fi


helm package $CHART_NAME
