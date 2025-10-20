### LitIt Platform Infrastructure Overview

#### **1. High-Level Architecture**

**Goal:** Build a unified, lightning-fast, AI-enhanced platform that merges all microservices into one scalable Go monolithic API architecture with modern tools for media, search, push, and real-time communication.

**Core Principles:**
- Monolithic Go Fiber backend
- Scalable, multi-tier infrastructure
- Mobile-first performance
- Real-time user interactions
- Global CDN acceleration

---

### **2. System Diagram (Logical Overview)**

```
+------------------------------+                   +------------------------------+
|          Mobile Apps         |                   |       Web Admin Portal       |
| Android / iOS / PWA          |                   | React / Next.js (Admin)      |
| - LitIt App API calls        |                   | - Management APIs            |
| - WebSockets (/ws)           |                   | - RPC → REST Migration       |
+---------------+--------------+                   +---------------+--------------+
                |                                              |
                +----------------------+-----------------------+
                                       |
                               +-------v--------+
                               | Caddy / Nginx  |   (HTTP/2, TLS, Brotli)
                               | SSL Termination|   Reverse Proxy Layer
                               +-------+--------+
                                       |
                           +-----------v-----------+
                           |   Go Fiber Monolith   |
                           |   REST API + WS Hub   |
                           |   Modular Internal PKG|
                           +-----------+-----------+
                                       |
        +--------------+---------------+--------------+--------------+
        |              |               |              |              |
+-------v-----+ +------v------+ +-------v------+ +-----v-------+ +----v-----+
| PostgreSQL  | | Redis Cache | | Meilisearch  | | Gotify WS  | | Bunny CDN|
| Main DB     | | Sessions,   | | Search Index | | Push/Alerts | | Images/VOD|
| 256GB RAM   | | Rate Limits | |              | | Channels    | | Hosting   |
+-------------+ +-------------+ +--------------+ +--------------+ +-----------+
```

---

### **3. Infrastructure Breakdown**

#### **Compute Layer**
- **API Server (64 GB RAM)**: Hosts Fiber monolith + Gotify + Caddy.
- **DB Server (256 GB RAM)**: Hosts Postgres, Redis, Meilisearch.

#### **Storage Layer**
- **Postgres 16:** Primary DB (with future read replica).
- **Redis 7:** Session cache, rate limiter, counters.
- **Meilisearch:** AI-powered fast search, content discovery.
- **Bunny.net:** CDN + VOD + object storage for images/videos.

#### **Networking Layer**
- HTTP/2 & WebSockets via Caddy/Nginx.
- Bunny CDN for static/media acceleration.

#### **CI/CD Layer**
- GitHub Actions → Build (Docker) → Push to GHCR → Portainer Deploy.
- Auto-deploy (dev/stg/prod) via Watchtower or SSH pipeline.

---

### **4. Data & Message Flow**

**User Request:**
1. Mobile app → HTTPS (Caddy/Nginx) → Fiber Monolith
2. Fiber authenticates (JWT + Device-ID)
3. Query hits Redis (cache) → fallback to Postgres → optional Meilisearch
4. Response → Gzip/Brotli compression → back to client

**Realtime Updates:**
1. App connects to `wss://mono.trmx.cc/ws`
2. Gotify (or Fiber WS hub) manages topic-based broadcasts
3. Redis Pub/Sub used for multi-node scalability

**Media Uploads:**
1. App uploads → Bunny Storage via signed URLs
2. CDN pull → `cdn.trmx.cc`

**Admin Management:**
- Web Admin Portal → REST `/admin/*` routes → Fiber Monolith
- Analytics, moderation, and content workflow APIs

---

### **5. Environments**

| Environment | Hostname             | Purpose             |
|--------------|----------------------|---------------------|
| Dev          | mono-dev.trmx.cc     | QA & integration    |
| Staging      | mono-stg.trmx.cc     | Pre-production      |
| Production   | mono.trmx.cc         | Live environment    |
| CDN          | cdn.trmx.cc          | Bunny edge delivery |

---

### **6. Performance & Optimization Stack**

| Layer | Tool | Function |
|-------|------|-----------|
| HTTP  | Caddy/Nginx | HTTP/2, Brotli, TLS auto-renew |
| App   | Go Fiber | REST API, WebSocket, Middleware |
| Cache | Redis | Hot data cache, rate limit, presence |
| DB    | Postgres | Primary data store, ACID transactions |
| Search| Meilisearch | Fast, fuzzy search engine |
| Push  | Gotify | Real-time push & topics |
| CDN   | Bunny | Video, image, and static media acceleration |

---

### **7. Security Layers**

- JWT (HS256/ES256) auth
- Device binding for mobile auth
- Role-based access for `/admin/*`
- Secret scanning + env-only sensitive keys
- TLS via Caddy auto-renew
- Optional Cloudflare WAF for DDOS

---

### **8. Marketing & Analytics Stack**

| Tool / Platform | Purpose |
|------------------|----------|
| Google Analytics / Firebase | Mobile engagement tracking |
| Meilisearch Logs | Search trends and popular content |
| Bunny Analytics | Media view statistics |
| Redis Counters | Internal metrics for active sessions |
| Prometheus / Grafana | Server performance monitoring |

---

### **9. Roadmap (Phased Plan)**

#### **Phase 1 (Week 1-2) — Foundations**
- Deploy Caddy + Fiber Monolith (Dev)
- Set up Postgres, Redis, Meili, Gotify, Bunny storage
- Import all microservices into `internal/*`

#### **Phase 2 (Week 3-4) — Data Migration**
- Migrate Scylla data → Postgres
- Sync Kafka/consumer features into Go routines
- Build Redis/Meili cache indexers

#### **Phase 3 (Week 5-6) — Real-time Layer**
- Deploy Gotify + Redis Pub/Sub bridge
- Integrate push channels into app (`/ws`)

#### **Phase 4 (Week 7-8) — Admin + Search + Optimization**
- Add `/admin/*` routes
- Optimize SQL indexes, Redis TTLs
- Implement Meili search autosync

#### **Phase 5 (Week 9+) — Observability + Scale**
- Deploy Prometheus, Grafana, Loki
- Add load balancer & read replica for Postgres
- CDN optimizations + Bunny analytics integration

---

### **10. Deliverables**

- `deploy/` folder: Docker Compose (API + DB)
- `internal/common/`: config, db, redis, logger modules
- `internal/ws/`: Gotify bridge + fallback WS hub
- `docs/routes.md`: all API endpoints (v1 & admin)
- `docs/openapi.yaml`: generated OpenAPI 3.0 schema
- `infra/diagram.drawio`: editable infrastructure diagram

---

Would you like me to generate the **editable `.drawio` file** next (with all nodes pre-labeled: API, DB, CDN, Redis, etc.) so you can open and modify it in diagrams.net or Lucidchart?

