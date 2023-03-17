export interface NavItem {
    text: string,
    route: string,
    icon: JSX.Element,
    page: JSX.Element,
    noAuth?: boolean
}
