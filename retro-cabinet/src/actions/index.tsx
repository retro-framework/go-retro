import * as constants from '../constants'

export interface ISetServerURL {
    type: constants.SET_SERVER_URL;
    payload: string,
}

export interface ISetSelectedHeadRef {
    payload: string,
    type: constants.SET_SELECTED_HEAD_REF;
}

export function setServerURL(newURL: string): ISetServerURL {
    return {
        payload: newURL,
        type: constants.SET_SERVER_URL,
    }
}

export function setSelectedHeadRef(newSelectedHeadRef: string): ISetSelectedHeadRef {
    return {
        payload: newSelectedHeadRef,
        type: constants.SET_SELECTED_HEAD_REF,
    }
}