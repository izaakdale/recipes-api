run:
	CLIENT_ID="H3Z6hHdkTCjn1RPJkxpgd89nxlzEEoVf" \
	CLIENT_SECRET="p0D64-cigAiiPIu6TecvVoHtTRVPyNCbBWULWzrmHzxEX2-UPSjqLL2i7A0V-l7c" \
	AUTH0_DOMAIN=dev-np0m9-q3.us.auth0.com \
	AUTH0_API_IDENTIFIER="https://api.recipes.io" \
	MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin" \
	MONGO_DATABASE=demo \
	go run .

swag:
	swagger generate spec -o ./swagger.yaml