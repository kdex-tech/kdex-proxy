declare class AppRouteItem {
    private config;
    constructor(config: {
        host?: HTMLElement;
        id: string;
        path: string;
    });
    host?: HTMLElement;
    id: string;
    path: string;
}
declare class AppRouteRegistry {
    private items;
    private callbacks;
    readonly pathSeparator: string;
    private history;
    constructor();
    _doNavigation(): void;
    _resetNavigationLinks(): void;
    addItem(item: AppRouteItem): void;
    basepath(): string;
    currentRoutepath(): AppRouteItem | null;
    findItem(id: string, path: string): AppRouteItem | null;
    getItems(): AppRouteItem[];
    navigate(location: string): void;
    onItemAdded(callback: (items: AppRouteItem[]) => void): void;
    registerRoutes(host: HTMLElement, ...paths: string[]): void;
    removeItem(item: AppRouteItem): void;
}
declare const appRouteRegistry: AppRouteRegistry;
export { AppRouteItem, AppRouteRegistry, appRouteRegistry };
//# sourceMappingURL=app-route.d.ts.map