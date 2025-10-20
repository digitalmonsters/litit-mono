# LitIt Infrastructure Visualization

Interactive, dual-theme (Tropical × Midnight) presentation of the LitIt Platform architecture.

## View Online
👉 https://digitalmonsters.github.io/litit-mono/presentation.html

## Edit or Extend
Each diagram lives in `/docs/infra/*.mmd`.
Edit with any Mermaid-compatible editor or insert via [draw.io](https://app.diagrams.net) → Arrange › Insert › Mermaid.

## Auto-Rendering
GitHub Actions (`.github/workflows/render-beautiful-diagrams.yml`) automatically regenerates high-definition SVGs into `docs/infra/out/`.

## Deploy
1. Push to `main`.  
2. Go to **Settings → Pages → Deploy from branch → main**.  
3. Your presentation will appear at <https://digitalmonsters.github.io/litit-mono/presentation.html>.

*(Optional)* Add a `CNAME` file in `/docs` to map a custom domain like `digitalmonsters.com/litit-mono`.
