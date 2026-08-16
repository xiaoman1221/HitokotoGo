function formatNumber(n) {
    const num = Number(n);
    return Number.isFinite(num) ? num.toLocaleString() : "-";
}

function escapeHtml(str) {
    return String(str)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

function $(id) {
    const el = document.getElementById(id);
    if (!el) console.warn("Element #" + id + " not found");
    return el;
}

async function loadStats() {
    const list = $("category-list");

    try {
        const response = await fetch("/stats/data", {
            method: "GET",
            cache: "no-store"
        });

        if (!response.ok) {
            throw new Error("请求失败: " + response.status);
        }

        const data = await response.json();

        const totalNumber = $("total-number");
        if (totalNumber) totalNumber.textContent = formatNumber(data.total);

        const bv = $("bundle-version");
        if (bv) bv.textContent = data.bundle_version || "-";

        const tq = $("total-queries");
        if (tq) tq.textContent = formatNumber(data.total_queries);

        const rpm = $("rpm");
        if (rpm) rpm.textContent = formatNumber(data.rpm);

        const loadEl = $("load");
        if (loadEl) {
            const l1 = data.load_1 != null ? data.load_1.toFixed(2) : "-";
            const l5 = data.load_5 != null ? data.load_5.toFixed(2) : "-";
            const l15 = data.load_15 != null ? data.load_15.toFixed(2) : "-";
            loadEl.textContent = l1 + " / " + l5 + " / " + l15;
        }

        const mem = $("memory");
        if (mem) mem.textContent = (data.memory_mb != null ? data.memory_mb.toFixed(1) : "-") + " MB";

        if (list && Array.isArray(data.categories)) {
            const countEl = $("category-count");
            if (countEl) countEl.textContent = data.categories.length;

            list.innerHTML = "";
            data.categories.forEach(cat => {
                const row = document.createElement("div");
                row.className = "category-row";
                row.innerHTML = `
                    <span class="category-key">${escapeHtml(cat.key)}</span>
                    <div class="category-main">
                        <span class="category-name">${escapeHtml(cat.name)}</span>
                        <span class="category-desc">${escapeHtml(cat.desc)}</span>
                    </div>
                    <span class="category-count">${formatNumber(cat.count)}</span>
                `;
                list.appendChild(row);
            });
        }
    } catch (error) {
        console.error("加载统计数据失败:", error);
        if (list) list.innerHTML = `<p class="error-msg">加载失败: ${escapeHtml(error.message)}</p>`;
    }
}

loadStats();
setInterval(loadStats, 60000);
