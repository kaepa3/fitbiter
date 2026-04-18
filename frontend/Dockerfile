# --- 基盤ステージ ---
FROM node:22-slim AS base
WORKDIR /app
COPY package*.json ./
# node_modulesをマウントではなくコンテナ内に閉鎖してインストール
RUN npm install

# --- 開発用ステージ ---
FROM base AS dev
COPY . .
EXPOSE 5173
# ホットリロードを効かせるため --host 指定
CMD ["npm", "run", "dev", "--", "--host"]

# --- ビルド用ステージ（成果物作成） ---
FROM base AS build-stage
COPY . .
RUN npm run build

# --- 本番用ステージ（Nginxで配信） ---
FROM nginx:alpine AS production
COPY --from=build-stage /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
