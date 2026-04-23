# ---------- build stage ----------
FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable

# cache dependencies separately
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# copy source code and build
COPY . .
COPY .env.local.docker .env.production
RUN pnpm build

# Isolate only production dependencies into /pruned
RUN pnpm deploy --filter . --prod --legacy /pruned

# ---------- runtime stage ----------
FROM node:22-alpine AS runner

WORKDIR /app

# RUN corepack enable

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# COPY --from=builder /app/build ./build
# COPY --from=builder /app/package.json ./
# COPY --from=builder /app/pnpm-lock.yaml ./

# Only copy the essential production files
COPY --from=builder /pruned/node_modules ./node_modules
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./

# RUN pnpm install --prod --frozen-lockfile

USER appuser

ENV NODE_ENV=production
EXPOSE 3000

CMD ["node", "build"]