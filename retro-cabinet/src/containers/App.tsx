import { connect, Dispatch } from 'react-redux';
import * as actions from '../actions/';
import RefSelector from '../RefSelector';
import { IStoreState } from '../types/index';

interface IStateFromProps {
  serverURL: string;
  selectedHeadRef: string;
}

interface IDispatchFromProps {
  setSelectedHeadRef: (_: string) => actions.ISetSelectedHeadRef;
  setServerURL: (_: string) => actions.ISetServerURL;
}

export function mapStateToProps(state: IStoreState): IStateFromProps {
  return state;
}

type IActions = actions.ISetSelectedHeadRef | actions.ISetServerURL;

function mapDispatchToProps(dispatch: Dispatch<IActions>): IDispatchFromProps {
  return {
    setSelectedHeadRef: (newHeadRef: string) => dispatch(actions.setSelectedHeadRef(newHeadRef)),
    setServerURL: (newServerURL: string) => dispatch(actions.setServerURL(newServerURL)),
  }
}

export default connect<IStateFromProps, IDispatchFromProps, void>(mapStateToProps, mapDispatchToProps)(RefSelector);