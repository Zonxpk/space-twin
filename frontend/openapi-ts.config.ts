import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  client: "@hey-api/client-fetch",
  input: "http://localhost:8080/swagger.yaml",
  output: "src/client",
  plugins: ["@tanstack/vue-query"],
});
