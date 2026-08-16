const backgroundApi = (typeof BACKGROUND_API !== "undefined" && BACKGROUND_API) || "https://t.alcy.cc/pc";
const quoteTextEl = document.getElementById("quote-text");
const quoteMetaEl = document.getElementById("quote-meta");
const bottomTipEl = document.getElementById("bottom-tip");

let bgActive = false;

/* ---------- 背景图：加载 + 明暗采样 ---------- */

const bgCanvas = document.createElement("canvas");
bgCanvas.width = 32;
bgCanvas.height = 32;
const bgCtx = bgCanvas.getContext("2d");

function loadImage(src, withCORS) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        if (withCORS) img.crossOrigin = "anonymous";
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error("load failed"));
        img.src = src;
    });
}

// 采样背景图平均亮度；CORS 受限或失败时返回 null（保持中性默认）
async function sampleBrightness(src) {
    try {
        let img;
        try {
            img = await loadImage(src, true);
        } catch {
            img = await loadImage(src, false);
        }
        bgCtx.drawImage(img, 0, 0, 32, 32);
        const data = bgCtx.getImageData(0, 0, 32, 32).data;
        let sum = 0;
        for (let i = 0; i < data.length; i += 4) {
            sum += 0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
        }
        return sum / (data.length / 4);
    } catch (err) {
        return null;
    }
}

async function applyBackground(url) {
    const src = `${url}?t=${Date.now()}`;
    document.body.style.backgroundImage = `url("${src}")`;
    document.body.classList.add("has-bg");

    // 根据亮度切换文字主题：亮图深字 / 暗图白字
    const brightness = await sampleBrightness(src);
    document.body.classList.remove("text-on-light", "text-on-dark");
    if (brightness == null) {
        return; // 无法采样：使用中性默认主题
    }
    if (brightness >= 140) {
        document.body.classList.add("text-on-light");
    } else {
        document.body.classList.add("text-on-dark");
    }
}

function refreshBackground() {
    applyBackground(backgroundApi);
}

/* ---------- 句子处理 ---------- */

function normalizeSentence(data) {
    const text =
        data.hitokoto ||
        data.text ||
        data.sentence ||
        data.content ||
        "此刻无言，静待下一句。";

    const fromParts = [data.from, data.from_who, data.creator].filter(Boolean);

    return {
        text,
        meta: fromParts.length > 0 ? "—— " + fromParts.join(" / ") : "—— HitokotoGo"
    };
}

/* ============ 根据句子长度自适应 ============ */

// 字号缩放系数：句子越长，字号越小（配合 CSS 变量 --quote-scale）
function textScale(length) {
    if (length <= 12) return 1;
    if (length <= 24) return 0.85;
    if (length <= 40) return 0.72;
    if (length <= 64) return 0.62;
    return 0.55;
}

// 行高：长句行高略增，保证可读性
function textLineHeight(scale) {
    return 1.22 + (1 - scale) * 0.18;
}

// 打字速度（每字毫秒）：短句慢打更有分量，长句快打避免拖沓
function typingInterval(length) {
    if (length <= 12) return 20;
    if (length <= 28) return 14;
    if (length <= 48) return 10;
    return 8;
}

// 退格速度：长句退格稍快
function backspaceInterval(length) {
    return length > 40 ? 7 : 10;
}

// 停留时间：按阅读时长自适应，并限制在 [REFRESH_INTERVAL, 12s]
function dwellTime(length) {
    const reading = length * 280;
    return Math.min(Math.max(reading, REFRESH_INTERVAL), 12000);
}

// 把自适应结果应用到当前句子
function applySentenceScale(length) {
    const scale = textScale(length);
    quoteTextEl.style.setProperty("--quote-scale", String(scale));
    quoteTextEl.style.setProperty("--quote-lh", String(textLineHeight(scale)));
}

function updateTip(ms) {
    if (!bottomTipEl || NO_REFRESH) return;
    const seconds = Math.max(1, Math.round(ms / 1000));
    bottomTipEl.textContent = `每 ${seconds} 秒自动刷新一句`;
}

/* ============================================ */

function typeText(element, text, callback) {
    element.textContent = "";
    let i = 0;
    const interval = typingInterval(text.length);
    const timer = setInterval(() => {
        element.textContent += text[i];
        i++;
        if (i >= text.length) {
            clearInterval(timer);
            if (callback) callback();
        }
    }, interval);
}

function backspaceText(element, callback) {
    const text = element.textContent;
    if (!text) {
        if (callback) callback();
        return;
    }
    let i = text.length;
    const interval = backspaceInterval(text.length);
    const timer = setInterval(() => {
        i--;
        element.textContent = text.substring(0, i);
        if (i <= 0) {
            clearInterval(timer);
            if (callback) callback();
        }
    }, interval);
}

function displaySentence(sentence, callback) {
    applySentenceScale(sentence.text.length);
    if (BACKGROUND_REFRESH && bgActive) {
        refreshBackground();
    }
    quoteMetaEl.textContent = sentence.meta;
    quoteMetaEl.style.opacity = "1";
    typeText(quoteTextEl, sentence.text, callback);
}

let isLoading = false;

async function loadSentence() {
    if (isLoading) return;
    isLoading = true;
    try {
        const response = await fetch("/v2", {
            method: "GET",
            cache: "no-store"
        });

        if (!response.ok) {
            throw new Error("请求失败: " + response.status);
        }

        const data = await response.json();
        const sentence = normalizeSentence(data);

        quoteMetaEl.style.opacity = "0";
        backspaceText(quoteTextEl, () => {
            displaySentence(sentence, () => {
                const nextInterval = dwellTime(sentence.text.length);
                updateTip(nextInterval);
                refreshTimer = setTimeout(loadSentence, nextInterval);
            });
        });
    } catch (error) {
        quoteTextEl.textContent = "句子加载失败";
        quoteMetaEl.textContent = "请检查服务或稍后重试";
        console.error(error);
        if (typeof refreshTimer !== "undefined") clearTimeout(refreshTimer);
        refreshTimer = setTimeout(loadSentence, REFRESH_INTERVAL);
    } finally {
        isLoading = false;
    }
}

var bgBtn = document.getElementById("bg-refresh-btn");
if (bgBtn) {
    bgBtn.addEventListener("click", () => {
        bgActive = true;
        refreshBackground();
    });
}

// 首次加载即刷新背景图
bgActive = true;
refreshBackground();

if (INITIAL_SENTENCE && INITIAL_SENTENCE.uuid) {
    const sentence = normalizeSentence(INITIAL_SENTENCE);
    displaySentence(sentence);
} else if (NO_REFRESH) {
    quoteTextEl.textContent = "句子未找到";
    quoteMetaEl.textContent = "请检查 UUID 是否正确";
}

var refreshTimer;

if (NO_REFRESH) {
    if (bottomTipEl) bottomTipEl.style.display = "none";
} else {
    loadSentence();
}
