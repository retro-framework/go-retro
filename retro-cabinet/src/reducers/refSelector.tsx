import * as actions from '../actions';
import { REF_LIST_ENTRIES_AVAILABLE, REF_LIST_INVALIDATE, SET_SELECTED_HEAD_REF_HASH } from '../constants/index';
import { IStoreRefSelectorState } from '../types/index';

export function refSelectorReducer(state: IStoreRefSelectorState, action: actions.IRefListEntriesAvailable | actions.IInvalidateRefList | actions.ISetSelectedHeadRefHash ): IStoreRefSelectorState {
  switch (action.type) {
    case REF_LIST_ENTRIES_AVAILABLE:
      return { ...state, loading: false, refs: action.payload };
    case REF_LIST_INVALIDATE:
      return { ...state, loading: true, refs: [] };
    case SET_SELECTED_HEAD_REF_HASH:
      return { ...state, selectedHash: action.payload };
  }
  return { ...state };
}