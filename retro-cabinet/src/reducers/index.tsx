import { ISetSelectedHeadRef, ISetServerURL } from '../actions';
import { SET_SELECTED_HEAD_REF, SET_SERVER_URL } from '../constants/index';
import { IStoreState } from '../types/index';

export function enthusiasm(state: IStoreState, action: ISetSelectedHeadRef | ISetServerURL): IStoreState {
  switch (action.type) {
    case SET_SELECTED_HEAD_REF:
      return { ...state, selectedHeadRef: action.payload };
    case SET_SERVER_URL:
      return { ...state, serverURL: action.payload };
  }
  return state;
}