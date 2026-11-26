"use strict";

const API_PREFIX = "/api/v1";

const CONFIG = {
  API_BASE: "", // "" = тот же адрес, что и у страницы
  LOGIN_URL: `${API_PREFIX}/login`,
  REGISTER_URL: `${API_PREFIX}/register`,
  TASKS_URL: `${API_PREFIX}/tasks`,
  TASK_URL: (id) => `${API_PREFIX}/tasks/${id}`,
};

/* ---------- тексты ошибок ----------
   Сервер присылает машинный code. Текст для пользователя живёт здесь,
   поэтому его можно менять, не трогая сервер. */
const ERROR_TEXTS = {
  UNAUTHORIZED: "Нужно войти в аккаунт.",
  INVALID_TOKEN: "Сессия истекла. Войдите снова.",
  INVALID_ID: "Некорректный номер задачи.",
  TASK_NOT_FOUND: "Задача не найдена. Возможно, её уже удалили.",
  DUPLICATE_TITLE: "Задача с таким названием уже есть.",
  INVALID_JSON: "Не удалось отправить данные. Обновите страницу и попробуйте снова.",
  USERNAME_TAKEN: "Это имя уже занято. Выберите другое.",
  INVALID_CREDENTIALS: "Неверное имя или пароль.",
  INTERNAL_ERROR: "Ошибка на сервере. Попробуйте позже.",
  // Эти коды появятся на сервере позже. Названия должны совпасть с твоими.
  PAYLOAD_TOO_LARGE: "Слишком длинный текст. Сократите название или описание.",
  NOT_FOUND: "Такого адреса нет.",
  METHOD_NOT_ALLOWED: "Это действие не поддерживается.",
};

const FIELD_LABELS = {
  username: "Имя пользователя",
  password: "Пароль",
  title: "Название",
  description: "Описание",
};

const GENERIC_TEXT = "Что-то пошло не так. Попробуйте ещё раз.";
const NETWORK_TEXT = "Нет связи с сервером. Проверьте подключение.";

function messageFor(code, field, status) {
  if (code === "VALIDATION_ERROR") {
    const label = FIELD_LABELS[field];
    return label ? `Поле «${label}» пустое или заполнено неверно.` : "Проверьте введённые данные.";
  }
  if (ERROR_TEXTS[code]) return ERROR_TEXTS[code];
  return status >= 500 ? ERROR_TEXTS.INTERNAL_ERROR : GENERIC_TEXT;
}

class ApiError extends Error {
  constructor(message, { status = 0, code = "", field = "", handled = false } = {}) {
    super(message);
    this.status = status;
    this.code = code;
    this.field = field;
    this.handled = handled; // true: ошибка уже обработана (например, выход из аккаунта)
  }
}

const $ = (id) => document.getElementById(id);
const state = { tasks: [], filter: "all", editingId: null };
let editRefs = {}; // поля формы редактирования, пока она открыта

/* ---------- работа с API ---------- */
function getToken() { return localStorage.getItem("token"); }

async function api(path, { method = "GET", body } = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers.Authorization = "Bearer " + token;

  let res;
  try {
    res = await fetch(CONFIG.API_BASE + path, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiError(NETWORK_TEXT, { code: "NETWORK" });
  }

  // Читаем тело как JSON только если сервер сам так заявил (у 204 тела нет).
  const isJSON = (res.headers.get("Content-Type") || "").includes("application/json");
  let data = null;
  if (isJSON) {
    try { data = await res.json(); } catch { /* битое тело: останется null */ }
  }

  if (res.ok) return data;

  const code = (data && data.code) || "";
  const field = (data && data.field) || "";

  if (token && (code === "INVALID_TOKEN" || code === "UNAUTHORIZED")) {
    logout(ERROR_TEXTS.INVALID_TOKEN);
    throw new ApiError(ERROR_TEXTS.INVALID_TOKEN, { status: res.status, code, handled: true });
  }
  throw new ApiError(messageFor(code, field, res.status), { status: res.status, code, field });
}

/* ---------- показ ошибок ---------- */
function clearError(errEl, fields = {}) {
  errEl.textContent = "";
  for (const input of Object.values(fields)) input.removeAttribute("aria-invalid");
}

function showError(errEl, err, fields = {}) {
  if (err.handled) return;
  if (!(err instanceof ApiError)) console.error(err);
  errEl.textContent = err instanceof ApiError ? err.message : GENERIC_TEXT;
  const input = fields[err.field];
  if (input) {
    input.setAttribute("aria-invalid", "true");
    input.focus();
  }
}

// Красная рамка пропадает, как только человек начинает исправлять поле.
function watch(input) {
  input.addEventListener("input", () => input.removeAttribute("aria-invalid"));
}

/* ---------- экраны ---------- */
function show(screen) {
  $("auth-screen").hidden = screen !== "auth";
  $("app-screen").hidden = screen !== "app";
}

function logout(message = "") {
  localStorage.removeItem("token");
  state.tasks = [];
  state.editingId = null;
  $("task-error").textContent = "";
  $("auth-error").textContent = message;
  show("auth");
}

/* ---------- вход / регистрация ---------- */
async function auth(kind) {
  const fields = { username: $("username"), password: $("password") };
  clearError($("auth-error"), fields);
  const username = $("username").value.trim();
  const password = $("password").value;

  try {
    const url = kind === "register" ? CONFIG.REGISTER_URL : CONFIG.LOGIN_URL;
    const data = await api(url, { method: "POST", body: { username, password } });
    if (!data || !data.token) throw new ApiError(GENERIC_TEXT);
    localStorage.setItem("token", data.token);
    $("auth-form").reset();
    await enterApp();
  } catch (e) {
    showError($("auth-error"), e, fields);
  }
}

/* ---------- задачи ---------- */
async function loadTasks() {
  try {
    const data = await api(CONFIG.TASKS_URL);
    state.tasks = Array.isArray(data) ? data : (data && data.tasks) || [];
    render();
  } catch (e) {
    showError($("task-error"), e);
  }
}

async function enterApp() {
  show("app");
  await loadTasks();
}

function visibleTasks() {
  return state.tasks.filter((t) =>
    state.filter === "all" ? true : state.filter === "done" ? t.done : !t.done);
}

function el(tag, props = {}, ...children) {
  const node = document.createElement(tag);
  Object.assign(node, props);
  node.append(...children);
  return node;
}

function render() {
  const list = $("task-list");
  list.replaceChildren();
  editRefs = {};
  const tasks = visibleTasks();
  $("empty").hidden = tasks.length > 0;

  for (const t of tasks) {
    const li = el("li", { className: "task" + (t.done ? " done" : "") });
    const check = el("input", { type: "checkbox", checked: !!t.done, ariaLabel: "Выполнено" });
    check.addEventListener("change", () => saveTask(t, { done: check.checked }));

    if (state.editingId === t.id) {
      const title = el("input", { value: t.title, ariaLabel: "Название" });
      const desc = el("input", { value: t.description || "", ariaLabel: "Описание" });
      editRefs = { title, description: desc };
      watch(title);
      watch(desc);
      const save = el("button", { className: "primary", type: "button" }, "Сохранить");
      const cancel = el("button", { type: "button" }, "Отмена");
      save.addEventListener("click", () => saveTask(t, { title: title.value.trim(), description: desc.value.trim() }));
      cancel.addEventListener("click", () => {
        state.editingId = null;
        clearError($("task-error"));
        render();
      });
      const form = el("div", { className: "edit-form" }, title, desc, el("div", { className: "row" }, save, cancel));
      li.append(check, form);
    } else {
      const body = el("div", { className: "body" }, el("div", { className: "t-title" }, t.title));
      if (t.description) body.append(el("div", { className: "t-desc" }, t.description));
      const edit = el("button", { type: "button" }, "Изменить");
      const del = el("button", { type: "button", className: "del" }, "Удалить");
      edit.addEventListener("click", () => {
        state.editingId = t.id;
        clearError($("task-error"));
        render();
      });
      del.addEventListener("click", () => removeTask(t));
      li.append(check, body, el("div", { className: "actions" }, edit, del));
    }
    list.append(li);
  }
}

// true, если задача создана: форму очищаем только в этом случае.
async function createTask(title, description) {
  const fields = { title: $("title"), description: $("description") };
  clearError($("task-error"), fields);
  try {
    await api(CONFIG.TASKS_URL, { method: "POST", body: { title, description, done: false } });
    await loadTasks();
    return true;
  } catch (e) {
    showError($("task-error"), e, fields);
    return false;
  }
}

async function saveTask(task, changes) {
  clearError($("task-error"), editRefs);
  const updated = { title: task.title, description: task.description || "", done: !!task.done, ...changes };
  try {
    await api(CONFIG.TASK_URL(task.id), { method: "PUT", body: updated });
    state.editingId = null;
    await loadTasks();
  } catch (e) {
    showError($("task-error"), e, editRefs);
    if (e.handled) return;
    // Форму редактирования оставляем открытой, чтобы исправить ввод.
    // Во всех остальных случаях (галочка, исчезнувшая задача) сверяем список с сервером.
    if (state.editingId === null || e.code === "TASK_NOT_FOUND") {
      state.editingId = null;
      await loadTasks();
    }
  }
}

async function removeTask(task) {
  if (!confirm(`Удалить задачу «${task.title}»?`)) return;
  clearError($("task-error"));
  try {
    await api(CONFIG.TASK_URL(task.id), { method: "DELETE" });
    await loadTasks();
  } catch (e) {
    showError($("task-error"), e);
  }
}

/* ---------- события ---------- */
watch($("username"));
watch($("password"));
watch($("title"));
watch($("description"));

$("auth-form").addEventListener("submit", (e) => { e.preventDefault(); auth("login"); });
$("btn-register").addEventListener("click", () => auth("register"));
$("btn-logout").addEventListener("click", () => logout());

$("task-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const ok = await createTask($("title").value.trim(), $("description").value.trim());
  if (ok) {
    $("task-form").reset();
    $("title").focus();
  }
});

document.querySelectorAll(".filters button").forEach((b) =>
  b.addEventListener("click", () => {
    state.filter = b.dataset.filter;
    document.querySelectorAll(".filters button").forEach((x) => x.classList.toggle("active", x === b));
    render();
  }));

/* ---------- старт ---------- */
getToken() ? enterApp() : show("auth");