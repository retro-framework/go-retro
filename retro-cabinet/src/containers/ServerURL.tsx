import { connect, Dispatch } from 'react-redux';
import * as urljoin from 'url-join';
import * as actions from '../actions';
import * as constants from '../constants';
import { IStoreState } from '../types/index';

import ServerURL from '../components/ServerURL';

interface IPropsFromState {
    url: URL;
}

interface IPropsFromDispatch {
    refreshRefsFromURL: (_: URL) => void,
}

export type IServerURLProps = IPropsFromState | IPropsFromDispatch;

export function mapStateToProps(state: IStoreState): IPropsFromState {
    return { url: state.server.url };
}

function mapDispatchToProps(dispatch: Dispatch<actions.Any>): IPropsFromDispatch {
    return {
        refreshRefsFromURL: (newURL: URL) => {
            dispatch(actions.refListInvalidate());
            dispatch(actions.setServerURL(newURL));
            fetch(urljoin(newURL + "/ref/"))
                .then((response) => response.json())
                .then((refs) => {
                    dispatch(actions.refListEntriesAvailable(refs));
                    const setHeadRef = {
                        payload: refs[Object.keys(refs)[0]],
                        type: constants.SET_SELECTED_HEAD_REF_HASH,
                    } as actions.ISetSelectedHeadRefHash;
                    dispatch(setHeadRef);
                })
                .catch((err) => console.error)
        }
    }
}

export default connect<IPropsFromState, IPropsFromDispatch, void>(mapStateToProps, mapDispatchToProps)(ServerURL);