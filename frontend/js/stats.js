const backgroundApi = "https://t.alcy.cc/pc";

function refreshBackground() {
    document.body.style.backgroundImage = `url("${backgroundApi}?t=${Date.now()}")`;
}

async function loadStats() {
    const grid = document.getElementById("category-grid");
    const totalCard = document.getElementById("total-card");
    const totalNumber = document.getElementById("total-number");

    try {
        const response = await fetch("/stats/data", {
            method: "GET",
            cache: "no-store"
        });

        if (!response.ok) {
            throw new Error("请求失败: " + response.status);
        }

        const data = await response.json();

        totalNumber.textContent = data.total.toLocaleString();
        document.getElementById("bundle-version").textContent = data.bundle_version || "-";
        totalCard.style.display = "block";

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
    } catch (error) {
        grid.innerHTML = `<div class="error-msg">加载失败: ${error.message}</div>`;
        console.error(error);
    }
}

refreshBackground();
loadStats();
