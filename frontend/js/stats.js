const backgroundApi = "https://t.alcy.cc/pc";

function refreshBackground() {
    document.body.style.backgroundImage = `url("${backgroundApi}?t=${Date.now()}")`;
}

function formatNumber(n) {
    return Number(n).toLocaleString();
}

function $(id) {
    const el = document.getElementById(id);
    if (!el) console.warn("Element #" + id + " not found");
    return el;
}

async function loadStats() {
    const grid = $("category-grid");
    const totalCard = $("total-card");
    const totalNumber = $("total-number");

    try {
        const response = await fetch("/stats/data", {
            method: "GET",
            cache: "no-store"
        });

        if (!response.ok) {
            throw new Error("请求失败: " + response.status);
        }

        const data = await response.json();

        if (totalNumber) totalNumber.textContent = data.total.toLocaleString();
        const bv = $("bundle-version");
        if (bv) bv.textContent = data.bundle_version || "-";
        if (totalCard) totalCard.style.display = "block";
        const ss = $("status-section");
        if (ss) ss.style.display = "block";

        if (grid && Array.isArray(data.categories)) {
            grid.innerHTML = "";
            data.categories.forEach(cat => {
                const card = document.createElement("div");
                card.className = "category-card";
                card.innerHTML = `
                    <div class="cat-name">${cat.name}</div>
                    <span class="cat-key">${cat.key}</span>
                    <div class="cat-count">${cat.count.toLocaleString()}</div>
                    <div class="cat-desc">${cat.desc}</div>
                `;
                grid.appendChild(card);
            });
        }

        const tq = $("total-queries");
        if (tq) tq.textContent = formatNumber(data.total_queries);
        const ld = $("load");
        if (ld) ld.textContent = (data.load_1 != null ? data.load_1.toFixed(2) : "-") + " " + (data.load_5 != null ? data.load_5.toFixed(2) : "-") + " " + (data.load_15 != null ? data.load_15.toFixed(2) : "-");
        const rpm = $("rpm");
        if (rpm) rpm.textContent = formatNumber(data.rpm);
        const mem = $("memory");
        if (mem) mem.textContent = (data.memory_mb != null ? data.memory_mb.toFixed(1) : "-") + " MB";
    } catch (error) {
        console.error("加载统计数据失败:", error);
        if (grid) grid.innerHTML = `<div class="error-msg">加载失败: ${error.message}</div>`;
    }
}

refreshBackground();
loadStats();
setInterval(loadStats, 60000);
