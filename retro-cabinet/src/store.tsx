import { applyMiddleware, combineReducers, compose, createStore } from 'redux';
import DevTools from './containers/DevTools';
import ObjectDatabaseViewerReducer from './reducers/ObjectDatabaseViewer';
import RefSelectorReducer from './reducers/RefSelector';
import ServerURLReducer from './reducers/ServerURL';
import IStoreState from './types/store';

function logger({ getState }: any) {
  return (next: any): any => (action: any): any => {
    // console.log('will dispatch', action)
    // Call the next dispatch method in the middleware chain.
    const returnValue = next(action)
    // console.log('state after dispatch', getState())
    // This will likely be the action itself, unless
    // a middleware further in chain changed it.
    return returnValue
  }
}

const rootReducer = combineReducers({
  odbViewer: ObjectDatabaseViewerReducer,
  refSelector: RefSelectorReducer,
  server: ServerURLReducer,
}) as any; // TODO: as any is a hack....

const enhancer = compose(
  applyMiddleware(logger),
  DevTools.instrument() // todo this probably shouldn't be included in live
);

const initialState = {};

export default createStore<IStoreState, any, any, any>(
  rootReducer,
  initialState,
  enhancer,
);