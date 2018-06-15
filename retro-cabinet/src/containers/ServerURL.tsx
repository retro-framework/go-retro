import { connect, Dispatch } from 'react-redux';
import * as urljoin from 'url-join';
import * as actions from '../actions';
import ServerURL from '../components/ServerURL';
import { IPropsFromDispatch, IPropsFromState } from '../types/ServerURL';
import IStoreState from '../types/store';

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
                })
                .catch((err) => console.error)
        }
    }
}

export default connect<IPropsFromState, IPropsFromDispatch>(mapStateToProps, mapDispatchToProps)(ServerURL);