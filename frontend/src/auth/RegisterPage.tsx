import React, { useState } from 'react';
import { Link } from 'react-router-dom';

// When redirected to login can use ?referrer=/somePage to redirect after successful login attempt

function RegisterForm() {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [email, setEmail] = useState("")
  const [remember, setRemember] = useState(false)
  
  return (
    <form method="post" action="/register#post">
      <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text" required={true} value={username} onChange={(event) => setUsername(event.target.value)}/>
      </div>
      <div className='formGroup'>
        <label htmlFor="password">Password</label>
        <input name="password" type="password" required={true} value={password} onChange={(event) => setPassword(event.target.value)}/>
      </div>
      <div className='formGroup'>
        <label htmlFor="email">Email (Optional - For password reset)</label>
        <input name="email" type="email" required={false} value={email} onChange={(event) => setEmail(event.target.value)}/>
      </div>
      <button className='signInButton'>REGISTER</button>
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
        <Link to="/register">Register</Link>
        <Link to="/resetPassword" style={{marginLeft: "auto"}}>Password Reset</Link>
      </div>
    </div>
  )
}

export function RegisterPage() {
  return (
    <div className="registerTile">
      <h1>Register</h1>
      <RegisterForm />
      <LoginOptions />
    </div>
  )
}