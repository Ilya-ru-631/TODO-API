"use strict";

const API_PREFIX = "/api/v1";

const CONFIG = {
  API_BASE: "",                  // "" = тот же адрес, что и страница. Иначе например "http://localhost:8080"

  LOGIN_URL: API_PREFIX + "/login",           // POST {username, password} -> токен
  REGISTER_URL: API_PREFIX +"/register",     // POST {username, password}
  TASKS_URL: API_PREFIX + "/tasks",           // GET список, POST создание

  TASK_URL: (id) => `${API_PREFIX}/tasks/${id}`, // GET / PUT / DELETE одной задачи
  
  TOKEN_FIELDS: ["token", "access_token", "jwt"], // в каком поле ответа /login лежит токен
  UPDATE_METHOD: "PUT",          
};
/* ============================================================ */

const $ = (id) => document.getElementById(id);
const state = { tasks: [], filter: "all", editingId: null };

/* ---------- работа с API ---------- */
function getToken() { return localStorage.getItem("token"); }

async function api(path, { method = "GET", body } = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers["Authorization"] = "Bearer " + token;

  const res = await fetch(CONFIG.API_BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }

  if (res.status === 401 && token) { logout(); throw new Error("Сессия истекла. Войдите снова."); }
  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || (typeof data === "string" && data) || `Ошибка ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

/* ---------- экраны ---------- */
function show(screen) {
  $("auth-screen").hidden = screen !== "auth";
  $("app-screen").hidden = screen !== "app";
}

function logout() {
  localStorage.removeItem("token");
  state.tasks = [];
  show("auth");
}

/* ---------- вход / регистрация ---------- */
async function auth(kind) {
  $("auth-error").textContent = "";
  const username = $("username").value.trim();
  const password = $("password").value;
  if (!username || !password) { $("auth-error").textContent = "Введите имя и пароль."; return; }

  try {
    if (kind === "register") {
      await api(CONFIG.REGISTER_URL, { method: "POST", body: { username, password } });
    }
    const data = await api(CONFIG.LOGIN_URL, { method: "POST", body: { username, password } });
    const token = data && CONFIG.TOKEN_FIELDS.map((f) => data[f]).find(Boolean);
    if (!token) throw new Error("Сервер не вернул токен. Проверь TOKEN_FIELDS в app.js.");
    localStorage.setItem("token", token);
    $("auth-form").reset();
    await enterApp();
  } catch (e) {
    $("auth-error").textContent = e.message;
  }
}

/* ---------- задачи ---------- */
async function loadTasks() {
  $("task-error").textContent = "";
  try {
    const data = await api(CONFIG.TASKS_URL);
    state.tasks = Array.isArray(data) ? data : (data && data.tasks) || [];
    render();
  } catch (e) { $("task-error").textContent = e.message; }
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
  const tasks = visibleTasks();
  $("empty").hidden = tasks.length > 0;

  for (const t of tasks) {
    const li = el("li", { className: "task" + (t.done ? " done" : "") });
    const check = el("input", { type: "checkbox", checked: !!t.done, ariaLabel: "Выполнено" });
    check.addEventListener("change", () => saveTask(t, { done: check.checked }));

    if (state.editingId === t.id) {
      const title = el("input", { value: t.title, ariaLabel: "Название" });
      const desc = el("input", { value: t.description || "", ariaLabel: "Описание" });
      const save = el("button", { className: "primary", type: "button" }, "Сохранить");
      const cancel = el("button", { type: "button" }, "Отмена");
      save.addEventListener("click", () => saveTask(t, { title: title.value.trim(), description: desc.value.trim() }));
      cancel.addEventListener("click", () => { state.editingId = null; render(); });
      const form = el("div", { className: "edit-form" }, title, desc, el("div", { className: "row" }, save, cancel));
      li.append(check, form);
    } else {
      const body = el("div", { className: "body" }, el("div", { className: "t-title" }, t.title));
      if (t.description) body.append(el("div", { className: "t-desc" }, t.description));
      const edit = el("button", { type: "button" }, "Изменить");
      const del = el("button", { type: "button", className: "del" }, "Удалить");
      edit.addEventListener("click", () => { state.editingId = t.id; render(); });
      del.addEventListener("click", () => removeTask(t));
      li.append(check, body, el("div", { className: "actions" }, edit, del));
    }
    list.append(li);
  }
}

async function createTask(title, description) {
  $("task-error").textContent = "";
  try {
    await api(CONFIG.TASKS_URL, { method: "POST", body: { title, description, done: false } });
    await loadTasks();   // перечитываем список, чтобы не гадать формат ответа
  } catch (e) { $("task-error").textContent = e.message; }
}

async function saveTask(task, changes) {
  $("task-error").textContent = "";
  const updated = { title: task.title, description: task.description || "", done: !!task.done, ...changes };
  try {
    await api(CONFIG.TASK_URL(task.id), { method: CONFIG.UPDATE_METHOD, body: updated });
    state.editingId = null;
    await loadTasks();
  } catch (e) { $("task-error").textContent = e.message; await loadTasks(); }
}

async function removeTask(task) {
  if (!confirm(`Удалить задачу «${task.title}»?`)) return;
  $("task-error").textContent = "";
  try {
    await api(CONFIG.TASK_URL(task.id), { method: "DELETE" });
    await loadTasks();
  } catch (e) { $("task-error").textContent = e.message; }
}

/* ---------- события ---------- */
$("auth-form").addEventListener("submit", (e) => { e.preventDefault(); auth("login"); });
$("btn-register").addEventListener("click", () => auth("register"));
$("btn-logout").addEventListener("click", logout);

$("task-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const title = $("title").value.trim();
  if (!title) return;
  await createTask(title, $("description").value.trim());
  $("task-form").reset();
  $("title").focus();
});

document.querySelectorAll(".filters button").forEach((b) =>
  b.addEventListener("click", () => {
    state.filter = b.dataset.filter;
    document.querySelectorAll(".filters button").forEach((x) => x.classList.toggle("active", x === b));
    render();
  }));

/* ---------- старт ---------- */
getToken() ? enterApp() : show("auth");
