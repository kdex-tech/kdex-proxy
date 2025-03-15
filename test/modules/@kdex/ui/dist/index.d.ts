import { AppRouteItem, AppRouteRegistry, appRouteRegistry } from './app-route';
declare class AppContainerElement extends HTMLElement {
    private appContainerTemplate;
    constructor();
    static elementName(): string;
    connectedCallback(): void;
}
declare class AppElement extends HTMLElement {
    routePath: string | null;
    constructor();
    static get observedAttributes(): string[];
    attributeChangedCallback(name: string, oldValue: string | null, newValue: string | null): void;
    basepath(): string;
    connectedCallback(): void;
    navigate(path: string): void;
    registerRoutes(...paths: string[]): void;
}
declare global {
    interface HTMLElementTagNameMap {
        'kdex-ui-app-container': AppContainerElement;
    }
}
export { AppContainerElement, AppElement, AppRouteItem, AppRouteRegistry, appRouteRegistry };
//# sourceMappingURL=index.d.ts.map