import * as actions from '../actions';
import { SERVER_SET_URL } from '../constants/index';
import { IState } from '../types/ServerURL';

export default function(state: IState, action: actions.ISetServerURL): IState {
  switch (action.type) {
    case SERVER_SET_URL:
      return { ...state, url: action.payload };
  }
  return { ...state }; 
}