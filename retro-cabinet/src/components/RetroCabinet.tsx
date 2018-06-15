import * as React from 'react';
import DevTools from '../containers/DevTools';
import ObjectDatabaseViewer from '../containers/ObjectDatabaseViewer';
import RefSelector from '../containers/RefSelector';
import ServerURL from '../containers/ServerURL';
import store from '../store';
import IStoreState from '../types/store';

import { Provider } from 'react-redux';

import 'codemirror/lib/codemirror.css';

class App extends React.Component<any, IStoreState> {
  constructor(props: any) {
    super(props);
  }
  public render() {
    return (
      <Provider store={store}>
        <div className="Retro">
          <div className="Retro__ConnInfo">
            <ServerURL />
            <RefSelector />
          </div>
          <div className="Retro__Panel">
            <h2>Object Database</h2>
            <ObjectDatabaseViewer />
          </div>
          <div className="Retro__Panel">
            <h2>PanelTitle</h2>
          </div>
          <div className="Retro__Panel">
            <h2>Command Console</h2>
            <button>Execute Command</button>
          </div>
          <div className="Retro__Panel">
            <h2>Boilerplate Commands</h2>
            <ul>
              <li><a href="#">Create User</a></li>
              <li><a href="#">Show Profile</a></li>
              <li><a href="#">Recover Password</a></li>
              <li><a href="#">Start Session</a></li>
            </ul>
          </div>
          <div className="Retro__Panel">
            <h2>StoreJSON</h2>
          </div>
          <DevTools />
        </div>
      </Provider>
    );
  }
}

export default App;
