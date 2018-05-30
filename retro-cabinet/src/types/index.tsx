
export interface IStoreServerURLState {
    readonly url: URL
}

export interface IStoreRefHash {
    readonly hash: string
    readonly name: string
}

export interface IStoreRefSelectorState {
    readonly error: Error | undefined
    readonly loading: boolean
    readonly refs: IStoreRefHash[]
    readonly selectedHash: string
}

export interface IStoreState {
    readonly server: IStoreServerURLState
    readonly refSelector: IStoreRefSelectorState
}