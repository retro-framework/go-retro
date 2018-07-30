import { connect, Dispatch } from 'react-redux';
import * as actions from '../actions';
import ObjectDatabaseViewerComponent from '../components/ObjectDatabaseViewer';
import {IPropsFromDispatch, IPropsFromState} from '../types/ObjectDatabaseViewer';
import IStoreState from '../types/store';

export function mapStateToProps(state: IStoreState): IPropsFromState {
    return { 
        checkpoints: state.odbViewer.checkpoints,
        isLoading: false,
        selectedHeadRefHash: state.odbViewer.selectedHeadRefHash,
     };
}

function mapDispatchToProps(dispatch: Dispatch<actions.Any>): IPropsFromDispatch {
    return {
        changeSelectedCheckpoint: actions.changeSelectedODBVCheckpoint
     }
}

export default connect<IPropsFromState, IPropsFromDispatch, void>(mapStateToProps, mapDispatchToProps)(ObjectDatabaseViewerComponent);