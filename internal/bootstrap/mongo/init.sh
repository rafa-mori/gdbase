#!/usr/bin/env bash
set -euo pipefail

: "${MONGO_INITDB_ROOT_USERNAME:=root}"
: "${MONGO_INITDB_ROOT_PASSWORD:=example}"
: "${MONGO_INITDB_DATABASE:=kubexds}"
: "${MONGO_APP_USER:=kubexds}"
: "${MONGO_APP_PASSWORD:=kubexds_pass}"

echo "[init.sh] creating app user '${MONGO_APP_USER}' on db '${MONGO_INITDB_DATABASE}'..."

cat >/tmp/init-app-user.js <<'JS'
const dbName = process.env.MONGO_INITDB_DATABASE || "kubexds";
const appUser = process.env.MONGO_APP_USER || "kubexds";
const appPass = process.env.MONGO_APP_PASSWORD || "kubexds_pass";

const admin = db.getSiblingDB('admin');
try {
  admin.auth(process.env.MONGO_INITDB_ROOT_USERNAME || "root",
             process.env.MONGO_INITDB_ROOT_PASSWORD || "example");
} catch (e) {
  // some images don't require auth during init
}

const appdb = db.getSiblingDB(dbName);
try { appdb.dropUser(appUser); } catch (e) {}

appdb.createUser({
  user: appUser,
  pwd:  appPass,
  roles: [{ role: "readWrite", db: dbName }]
});

print(`[init.js] user '${appUser}' provisioned on db '${dbName}'`);
JS

mongosh --quiet --file /tmp/init-app-user.js
rm -f /tmp/init-app-user.js

echo "[init.sh] done."
