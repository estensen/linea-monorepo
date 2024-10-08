import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "(spec|test).ts$",
  verbose: true,
  globalSetup: "./env-setup/global-setup.ts",
  setupFilesAfterEnv: ["./env-setup/setup.ts"],
  globalTeardown: "./env-setup/global-teardown.ts",
};

export default config;
