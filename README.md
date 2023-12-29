# FFAAS

Fast functions as a service platform powered by WASM.

coming soon...

## API Server Endpoints

### /status

Get server status

- Method: `POST`
- Response Content-Type: `application/json`

Request Body: `empty`

Example Response:

```json
{
  "status": "ok"
}
```

---

### /application/\<id\>

Get Application by ID

- Method: `GET`
- Response Content-Type: `application/json`

- Request Body: `empty`

Example Response:

```json
{
  "id": "09248ef6-c401-4601-8928-5964d61f2c61",
  "name": "My first ffaas app",
  "endpoint": "http://0.0.0.0:4000/09248ef6-c401-4601-8928-5964d61f2c61",
  "active_deploy_id": "aeacab67-91d6-45c1-ae29-f27922b0fcf0",
  "deploy_history": [
    {
      "id": "aeacab67-91d6-45c1-ae29-f27922b0fcf0",
      "app_id": "09248ef6-c401-4601-8928-5964d61f2c61",
      "hash": "c4dd6753109e47b317a4fc792d231b64",
      "created_at": "2023-12-29T12:19:20.594726Z"
    }
  ],
  "created_at": "2023-12-29T12:19:20.574321Z"
}
```

---

### /application

Create a new application

- Method: `POST`
- Request Content-Type: `application/json`
- Response Content-Type: `application/json`

Example Request Body:

```json
{
  "name": "my-app"
}
```

Example Response Body:

```json
{
  "id": "2488b7be-e3d3-4e4c-8f79-13d9d568483d",
  "name": "my-app",
  "endpoint": "http://0.0.0.0:4000/2488b7be-e3d3-4e4c-8f79-13d9d568483d",
  "active_deploy_id": "00000000-0000-0000-0000-000000000000",
  "deploy_history": [],
  "created_at": "2023-12-29T12:08:20.542039Z"
}
```

---

### /application/\<id\>/deploy

Deploy Wasm Blob to Application

- Method: `POST`
- Request Content-Type: `application/octet-stream`
- Response Content-Type: `application/json`

Request Body: WASM file

Example Response:

```json
{
  "id": "e2a1ceea-d19e-4231-adc9-995ac61bdaf0",
  "app_id": "2488b7be-e3d3-4e4c-8f79-13d9d568483d",
  "hash": "75b196bcd44611d9f74d62ed16a54e03",
  "created_at": "2023-12-29T12:12:39.91252Z"
}
```

---

## Wasm Server Endpoints

### /\<id\>

Call the Wasm function

- Method: `ALL`
- Request Content-Type: `any`
- Response Content-Type: `any`

Request Body: `any` (passed to function)

Response Body: `any` (returned from function)
