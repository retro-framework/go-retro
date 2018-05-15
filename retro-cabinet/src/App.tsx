import * as React from 'react';
import * as CodeMirror from 'react-codemirror';
import { createStore } from 'redux';

import ObjectDatabaseViewer from './ObjectDatabaseViewer';
import RefSelector from './RefSelector';

import { enthusiasm } from './reducers/index';
import { IStoreState } from './types/index';

import 'codemirror/lib/codemirror.css';
import './App.css';
import './index.css';

interface IAppState {
  headRef: string;
  code: string;
}

const store = createStore<IStoreState, any, any, any>(enthusiasm, {
  selectedHeadRef: 'refs/heads/master',
  serverURL: "localhost:8080",
});

class App extends React.Component<any, IAppState> {
  constructor(props: any) {
    super(props);
    const placeholderCode = `{
      "path": "agg/123",
      "name": "dummyCmd",
      "args": {
          "str": "bar",
          "int": 123
      }
  }`
    this.state = { 
      code: placeholderCode,
      headRef: "", 
    };
  }
  public refChanged = (newRef: string) => {
    this.setState({ headRef: newRef });
  }
  public updateCode = (newCode: any) => {
    this.setState({ code: newCode });
  }
  public render() {
    const options = {
      lineNumbers: true,
      mode: { name: "javascript", json: true },
    };
    return (
      <div className="Retro">
        <div className="Retro__ConnInfo">
          <input type="text" defaultValue="localhost:8080" />
          <RefSelector onChange={this.refChanged} />
        </div>
        <div className="Retro__Panel">
          <h2>Object Database</h2>
          <ObjectDatabaseViewer headRef={this.state.headRef} />
        </div>
        <div className="Retro__Panel">
          <h2>PanelTitle</h2>
        </div>
        <div className="Retro__Panel">
          <h2>Command Console</h2>
          <CodeMirror value={this.state.code} onChange={this.updateCode} options={options} />
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
          <h2>PanelTitle</h2>
        </div>
      </div>
    );
  }
}

export default App;
