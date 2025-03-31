// eslint-disable-next-line @typescript-eslint/no-unused-vars
import React from "react";

export async function submitFormData(
  formURL: string,
  options?: RequestInit,
  callback?: (success: boolean, responseData?: unknown) => void,
  setLoading?: React.Dispatch<React.SetStateAction<boolean>>) {

  if (setLoading !== undefined) {
    setLoading(true)
  }

  const request = new Request(formURL, {
    method: "POST",
  })
  let isResponseOK = false // Used when await response.json() fails due to empty response
  let parsed = false

  try {

    const response = await fetch(request, options)
    isResponseOK = response.ok

    const responseData = await response.json()
    parsed = true
    if (callback !== undefined) {
      callback(isResponseOK, responseData)
    }

  } catch (e) {
    console.log(e)
    if (callback !== undefined && !parsed) {
      // Called when bad fetch or await response.json() fail
      callback(isResponseOK)
    }

  } finally {
    if (setLoading !== undefined) {
      setLoading(false)
    }
  }
}