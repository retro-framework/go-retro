export interface IStateLoadable {
    readonly loading: boolean
}

export interface IStateErrorable {
    readonly error?: Error | undefined
}

export interface IRefHash {
    readonly hash: string
    readonly name: string
}

export interface ICheckpoint {
    readonly hash: string
    readonly parentHashes: string[]
    readonly affixHash: string
    readonly commandDesc: string
    readonly summary: string
}