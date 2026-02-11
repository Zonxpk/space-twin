import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  client: "@hey-api/client-fetch",
  input: "http://localhost:8080/api/v1/openapi.yaml",
  output: "src/client",
  plugins: ["@tanstack/vue-query"],
});
