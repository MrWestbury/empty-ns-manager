$version="0.0.19"

docker build -t mortbury.azurecr.io/empty-ns-controller:${version} --build-arg "APPVERSION=${version}"  .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker build failed"
    exit $LASTEXITCODE
}

docker push mortbury.azurecr.io/empty-ns-controller:${version}
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker push failed"
    exit $LASTEXITCODE
}

helm upgrade --namespace empty-ns --create-namespace --install --set "image.tag=${version}" empty-ns-manager ./charts/empty-ns