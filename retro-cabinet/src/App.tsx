import * as React from 'react';

import RetroCabinet from './components/RetroCabinet';

import './App.css';
import './index.css';

class App extends React.Component {
  constructor(props: any) {
    super(props);
  }
  public render() {
    return (
      <RetroCabinet />
    );
  }
}

export default App;
