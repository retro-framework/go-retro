import { ISetServerURL } from '../actions';
import { SERVER_SET_URL } from '../constants/index';
import { IStoreServerURLState } from '../types/index';

export function serverURLReducer(state: IStoreServerURLState, action: ISetServerURL): IStoreServerURLState {
  switch (action.type) {
    case SERVER_SET_URL:
      return { ...state, url: action.payload };
  }
  return { ...state }; 
}