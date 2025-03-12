import React from 'react';

function LoginForm() {
  return (
    <form>
      <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text"/>
      </div>
      <div className='formGroup'>
        <label htmlFor="password">Password</label>
        <input name="password" type="text"/>
      </div>
      <button className='signInButton'>SIGN IN</button>
    </form>
  )
}

function LoginOptions() {
  return (
    <div>
      <label>
        <input type="checkbox" style={{marginLeft: "0"}}/>
        Keep me logged in
      </label>
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