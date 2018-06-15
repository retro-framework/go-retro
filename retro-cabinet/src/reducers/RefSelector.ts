import * as actions from '../actions';
import { REF_LIST_ENTRIES_AVAILABLE, REF_LIST_INVALIDATE, SET_SELECTED_HEAD_REF_HASH } from '../constants/index';
import { IState } from '../types/RefSelector';

export default function (state: IState, action: actions.IRefListEntriesAvailable | actions.IInvalidateRefList | actions.ISetSelectedHeadRefHash): IState {
  switch (action.type) {
    case REF_LIST_ENTRIES_AVAILABLE:
      return { ...state, loading: false, refs: action.payload, selectedHash: action.payload[action.payload.length -1].hash };
    case REF_LIST_INVALIDATE:
      return { ...state, loading: true, refs: [] };
    case SET_SELECTED_HEAD_REF_HASH:
      return { ...state, selectedHash: action.payload };
  }
  return { ...state };
}