curl -sL https://a01tools.blob.core.windows.net/droid/linux/a01droid -o tools/a01droid
chmod +x tools/a01droid
chmod +x tools/get_index

docker build -t azureclidev.azurecr.io/azuresdkforgo-test-azure:go1.10-test .
