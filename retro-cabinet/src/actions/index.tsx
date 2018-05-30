import * as constants from '../constants'
import * as types from '../types'

export interface IInvalidateRefList {
    type: constants.REF_LIST_INVALIDATE,
}

export interface IRefListEntriesAvailable {
    type: constants.REF_LIST_ENTRIES_AVAILABLE,
    payload: types.IStoreRefHash[],
}

export interface ISetServerURL {
    type: constants.SERVER_SET_URL;
    payload: URL,
}

export interface ISetSelectedHeadRefHash {
    payload: string,
    type: constants.SET_SELECTED_HEAD_REF_HASH;
}

export function refListInvalidate(): IInvalidateRefList {
    return { type: constants.REF_LIST_INVALIDATE };
}

export function setServerURL(newURL: URL): ISetServerURL {
    return { type: constants.SERVER_SET_URL, payload: newURL };
}

export function setSelectedHeadRefHash(newHeadRefHash: string): ISetSelectedHeadRefHash {
    return {
        payload: newHeadRefHash,
        type: constants.SET_SELECTED_HEAD_REF_HASH,
    }
}

export function refListEntriesAvailable(obj: any): IRefListEntriesAvailable {
    const refs: types.IStoreRefHash[] = [];
    for (const name in obj) {
        if (obj.hasOwnProperty(name)) {     
            refs.push({ name, hash: obj[name] });
        }
    }
    return {
        payload: refs,
        type: constants.REF_LIST_ENTRIES_AVAILABLE,
    }
}

// https://github.com/piotrwitek/react-redux-typescript-guide#rootaction---union-type-of-all-action-objects
export type Any =
    | IInvalidateRefList
    | IRefListEntriesAvailable
    | ISetSelectedHeadRefHash
    | ISetServerURL;