import { createBrowserHistory as r } from "history";
class o {
  constructor(t) {
    this.config = t, this.host = t.host, this.id = t.id, this.path = t.path;
  }
}
class h {
  constructor() {
    this.items = [], this.callbacks = [];
    const t = document.querySelector('html head meta[name="path-separator"]');
    this.pathSeparator = (t == null ? void 0 : t.getAttribute("content")) || "/_/", this.history = r(), this.history.listen(() => {
      this._doNavigation();
    }), document.addEventListener("DOMContentLoaded", this._resetNavigationLinks.bind(this)), document.addEventListener("DOMContentLoaded", this._doNavigation.bind(this));
  }
  _doNavigation() {
    const t = this.currentRoutepath(), e = new Set(this.getItems().map((i) => i.host).filter((i) => i !== void 0));
    for (const i of e)
      t && i.id === t.id ? i.setAttribute("route-path", `/${t.path}`) : i.setAttribute("route-path", "");
  }
  _resetNavigationLinks() {
    for (let t of document.querySelectorAll("a"))
      if (t.href.startsWith(document.location.origin + this.basepath())) {
        const e = new URL(t.href);
        t.onclick = () => (this.history.push(e.pathname), !1), t.href = "javascript:void(0)";
      }
  }
  addItem(t) {
    this.items.push(t), this.callbacks.forEach((e) => e(this.items));
  }
  basepath() {
    return window.location.pathname.includes(this.pathSeparator) ? window.location.pathname.split(this.pathSeparator, 2)[0] : window.location.pathname.endsWith("/") ? window.location.pathname.slice(0, -1) : window.location.pathname;
  }
  currentRoutepath() {
    if (window.location.pathname.includes(this.pathSeparator)) {
      const t = window.location.pathname.split(this.pathSeparator, 2)[1], [e, i] = t.split("/", 2);
      return new o({
        id: e,
        path: i
      });
    }
    return null;
  }
  findItem(t, e) {
    return this.items.find((i) => i.id === t && i.path === e) || null;
  }
  getItems() {
    return this.items;
  }
  navigate(t) {
    this.history.push(t);
  }
  onItemAdded(t) {
    this.callbacks.push(t);
  }
  registerRoutes(t, ...e) {
    for (const i of e)
      this.addItem(new o({
        host: t,
        id: t.id,
        path: i
      }));
  }
  removeItem(t) {
    this.items = this.items.filter((e) => e !== t), this.callbacks.forEach((e) => e(this.items));
  }
}
const a = new h();
class s extends HTMLElement {
  constructor() {
    super(), this.appContainerTemplate = document.createElement("template"), this.appContainerTemplate.innerHTML = `
    <slot><em>Application Container (placeholder)</em></slot>
    `;
  }
  static elementName() {
    return "kdex-ui-app-container";
  }
  connectedCallback() {
    this.attachShadow({ mode: "closed" }).appendChild(this.appContainerTemplate.content.cloneNode(!0));
  }
}
class l extends HTMLElement {
  constructor() {
    if (super(), this.routePath = null, !(this.parentElement instanceof s))
      throw new Error("Parent AppContainerElement not found");
  }
  static get observedAttributes() {
    return ["route-path"];
  }
  attributeChangedCallback(t, e, i) {
    t === "route-path" && (this.routePath = i), this.connectedCallback();
  }
  basepath() {
    return a.basepath();
  }
  connectedCallback() {
  }
  navigate(t) {
    t.startsWith("/") || (t = `/${t}`), a.navigate(`${this.basepath()}${a.pathSeparator}${this.id}${t}`);
  }
  registerRoutes(...t) {
    this.id && a.registerRoutes(this, ...t);
  }
}
customElements.get(s.elementName()) || customElements.define(s.elementName(), s);
export {
  s as AppContainerElement,
  l as AppElement,
  o as AppRouteItem,
  h as AppRouteRegistry,
  a as appRouteRegistry
};
//# sourceMappingURL=index.mjs.map
