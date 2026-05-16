const backgroundApi = "https://t.alcy.cc/pc";
const quoteCardEl = document.getElementById("quote-card");
const quoteTextEl = document.getElementById("quote-text");
const quoteMetaEl = document.getElementById("quote-meta");

function refreshBackground() {
    document.body.style.backgroundImage = `url("${backgroundApi}?t=${Date.now()}")`;
}

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

function setCardWidth() {
    quoteCardEl.style.width = "80vw";
}

function typeText(element, text, callback) {
    element.textContent = "";
    let i = 0;
    const timer = setInterval(() => {
        element.textContent += text[i];
        i++;
        if (i >= text.length) {
            clearInterval(timer);
            if (callback) callback();
        }
    }, 10);
}

function backspaceText(element, callback) {
    const text = element.textContent;
    if (!text) {
        if (callback) callback();
        return;
    }
    let i = text.length;
    const timer = setInterval(() => {
        i--;
        element.textContent = text.substring(0, i);
        if (i <= 0) {
            clearInterval(timer);
            if (callback) callback();
        }
    }, 10);
}

function displaySentence(sentence, callback) {
    setCardWidth();
    if (BACKGROUND_REFRESH) {
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
                const readingTime = sentence.text.length * 300;
                const nextInterval = Math.max(REFRESH_INTERVAL, readingTime);
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

setCardWidth();
refreshBackground();
var bgBtn = document.getElementById("bg-refresh-btn");
if (bgBtn) bgBtn.addEventListener("click", refreshBackground);

if (INITIAL_SENTENCE && INITIAL_SENTENCE.uuid) {
    const sentence = normalizeSentence(INITIAL_SENTENCE);
    displaySentence(sentence);
} else if (NO_REFRESH) {
    quoteTextEl.textContent = "句子未找到";
    quoteMetaEl.textContent = "请检查 UUID 是否正确";
}

var refreshTimer;

if (NO_REFRESH) {
    document.getElementById("bottom-tip").style.display = "none";
} else {
    loadSentence();
}
