import { connect, Dispatch } from 'react-redux';
import * as actions from '../actions/';
import RefSelector from '../components/RefSelector';
import { IStoreRefSelectorState, IStoreState } from '../types/index';

/* tslint:disable: no-console, no-empty-interface */

interface IStateFromProps { }

interface IDispatchFromProps { }

export function mapStateToProps(state: IStoreState): IStoreRefSelectorState {
    return { 
        error: undefined,
        loading: state.refSelector.loading,
        refs: state.refSelector.refs,
        selectedHash: (state.refSelector.selectedHash ? state.refSelector.selectedHash : "refs/heads/master"),
     };
}

function mapDispatchToProps(dispatch: Dispatch<actions.ISetSelectedHeadRefHash>): IDispatchFromProps {
    return { }
}

export default connect<IStateFromProps, IDispatchFromProps, void>(mapStateToProps, mapDispatchToProps)(RefSelector);