import { applyMiddleware, combineReducers, compose, createStore } from 'redux';

import DevTools from './containers/DevTools';

import { refSelectorReducer } from './reducers/refSelector';
import { serverURLReducer } from './reducers/serverURL';
import { IStoreState } from './types/index';

/* tslint:disable:no-console */

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
    refSelector: refSelectorReducer,
    server: serverURLReducer, 
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