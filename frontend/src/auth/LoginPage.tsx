import React, { useState } from 'react';

// When redirected to login can use ?referrer=/somePage to redirect after successful login attempt

function LoginForm() {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [remember, setRemember] = useState(false)
  
  return (
    <form method="post" action="/login#post">
      <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text" required={true} value={username} onChange={(event) => setUsername(event.target.value)}/>
      </div>
      <div className='formGroup'>
        <label htmlFor="password">Password</label>
        <input name="password" type="password" required={true} value={password} onChange={(event) => setPassword(event.target.value)}/>
      </div>
      <button className='signInButton'>SIGN IN</button>
      <label>
        <input type="checkbox" style={{marginLeft: "0"}} checked={remember} onChange={() => setRemember(!remember)}/>
        Keep me logged in
      </label>
    </form>
  )
}

function LoginOptions() {
  return (
    <div>
      <div className="loginOptions">
        <a href="#register">Register</a>
        <a href="#resetPassword" style={{marginLeft: "auto"}}>Password Reset</a>
      </div>
    </div>
  )
}

export function LoginPage() {
  return (
    <div className="loginTile">
      <h1>Sign In</h1>
      <LoginForm />
      <LoginOptions />
    </div>
  )
}