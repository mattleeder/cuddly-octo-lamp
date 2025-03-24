import React from 'react';
import './App.css';
import { MatchRoom } from './chess/MatchRoom.tsx';
import { QueueTiles } from './chess/QueueTiles.tsx';
import { LoginPage } from './auth/LoginPage.tsx';
import { TopNavBar } from './TopNavBar.tsx';
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import { HomepageMatch } from './chess/HomepageMatch.tsx';
import { PlayerInfoTile } from './PlayerInfoTile.tsx';
import { PageNotFound } from './PageNotFound.tsx';
import { RegisterPage } from './auth/RegisterPage.tsx';
import { AuthProvider } from './auth/AuthContext.tsx';
import { WatchPage } from './WatchPage.tsx';
import { AccountPage } from './auth/AccountPage.tsx';

function App() {
  console.log(import.meta.env.VITE_API_URL)
  console.log(`React Version: ${React.version}`)

  return (
    <>
      <AuthProvider>
        <PlayerInfoTile>
          <Router>
            <TopNavBar />
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/matchroom/:matchid" element={<MatchRoom />} />
              <Route path="/login" element={<LoginPage />}/>
              <Route path="/register" element={<RegisterPage/>}/>
              <Route path="/watch" element={<WatchPage />}/>
              <Route path="/user/:username" element={<AccountPage/>}/>
              <Route path="*" element={<PageNotFound />}/>
            </Routes>
          </Router>
        </PlayerInfoTile>
      </AuthProvider>
    </>
  )
}

function Home() {

  return (
    <>
      <QueueTiles />
      <HomepageMatch />
    </>
  )
}

export default App
