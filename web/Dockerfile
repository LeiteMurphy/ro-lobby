# syntax=docker/dockerfile:1
# Imagem do web (ADR-06, D-02 da spec home-local). O estágio final leva só o build do
# adapter-node e as dependências de produção, rodando como o usuário node (RN-05).

FROM mirror.gcr.io/library/node:24.14.0-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Só as dependências de produção, instaladas do zero (o npm prune deixa sobras).
FROM mirror.gcr.io/library/node:24.14.0-alpine AS prod-deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --ignore-scripts && mkdir -p node_modules

FROM mirror.gcr.io/library/node:24.14.0-alpine
WORKDIR /app
ENV NODE_ENV=production PORT=3000
COPY --from=build --chown=node:node /app/package.json ./
COPY --from=prod-deps --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/build ./build
USER node
EXPOSE 3000
HEALTHCHECK --interval=5s --timeout=3s --retries=12 \
  CMD node -e "fetch('http://127.0.0.1:'+(process.env.PORT||3000)+'/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"
CMD ["node", "build"]
