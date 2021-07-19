use std::os::raw;

mod seeds;

/// This file implements the low-level, C-compatible API exposed by alienlib.
/// These functions are mostly safe, at least to C standards.

/// Result type for alien_check_password.
#[repr(u32)]
pub enum PasswordResult {
    Okay = 0,
    InvalidPassword = 1,
    UsernameNull = 2,
    PasswordNull = 3,
    InvalidUnicode = 4,
    CouldNotLoadSeeds = 5,
    CouldNotParseSeeds = 6,
    NoSuchUser = 7,
}

/// Convert a PasswordResult into a const char* containing a human-readable error string. The
/// returned string must _not_ be freed using alien_string_free.
#[no_mangle]
pub extern "C" fn alien_password_result_string(result: PasswordResult) -> *const raw::c_char {
    let errstr: &'static str = match result {
        PasswordResult::Okay => "okay\0",
        PasswordResult::InvalidPassword => "invalid password\0",
        PasswordResult::UsernameNull => "given username is null\0",
        PasswordResult::PasswordNull => "given password is null\0",
        PasswordResult::InvalidUnicode => "given password contains invalid unicode\0",
        PasswordResult::CouldNotLoadSeeds => "could not load configuration (seeds)\0",
        PasswordResult::CouldNotParseSeeds => "could not parse configuration (seeds)\0",
        PasswordResult::NoSuchUser => "no such user\0",
    };
    errstr.as_ptr() as *const raw::c_char
}

/// Check a given username and password against the system database (seeds).
#[no_mangle]
pub extern "C" fn alien_check_password(username: *const raw::c_char, password: *const raw::c_char) -> PasswordResult {
    if username.is_null() {
        return PasswordResult::UsernameNull;
    }
    let cstr = unsafe { std::ffi::CStr::from_ptr(username) };
    let username = match cstr.to_str() {
        Ok(s) => s,
        Err(_) => return PasswordResult::InvalidUnicode,
    };

    if password.is_null() {
        return PasswordResult::PasswordNull;
    }
    let cstr = unsafe { std::ffi::CStr::from_ptr(password) };
    let password = match cstr.to_str() {
        Ok(s) => s,
        Err(_) => return PasswordResult::InvalidUnicode,
    };

    let path = "/etc/seeds/config_data.json";
    let contents = match std::fs::read_to_string(path) {
        Ok(c) => c,
        Err(_) => return PasswordResult::CouldNotLoadSeeds,
    };
    let seeds = match seeds::Seeds::parse(&contents) {
        Ok(s) => s,
        Err(_) => return PasswordResult::CouldNotParseSeeds,
    };

    let user = match seeds.get_admin(username) {
        Some(u) => u,
        None => return PasswordResult::NoSuchUser,
    };

    if user.check_password(password) {
        return PasswordResult::Okay;
    }
    return PasswordResult::InvalidPassword;
}

/// Return the SSH public key of a given user, as configured in the admin panel. Returns null on
/// error or if no public key has been set. The returned string must be freed using
/// alien_string_free.
/// 
/// Note: for this to work, the firmware must be modified and an 'SSHKey' field must be added to
/// 0ADM_PASSWORD in seeds. This is done by default by the default CFW built.
#[no_mangle]
pub extern "C" fn alien_get_ssh_pubkey(username: *const raw::c_char) -> *mut raw::c_char {
    if username.is_null() {
        return std::ptr::null_mut();
    }
    let cstr = unsafe { std::ffi::CStr::from_ptr(username) };
    let username = match cstr.to_str() {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let path = "/etc/seeds/config_data.json";
    let contents = match std::fs::read_to_string(path) {
        Ok(c) => c,
        Err(_) => return std::ptr::null_mut(),
    };
    let seeds = match seeds::Seeds::parse(&contents) {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let user = match seeds.get_admin(username) {
        Some(u) => u,
        None => return std::ptr::null_mut(),
    };
    return match &user.ssh_key {
        Some(k) => {
            let k = String::from(k);
            let k = match std::ffi::CString::new(k) {
                Ok(v) => v,
                Err(_) => return std::ptr::null_mut(),
            };
            k.into_raw()
        },
        None => std::ptr::null_mut(),
    }

}

/// Free a string, as returned by alien_get_ssh_pubkey.
#[no_mangle]
pub extern "C" fn alien_string_free(s: *mut raw::c_char) {
    unsafe {
        if s.is_null() {
            return;
        }
        std::ffi::CString::from_raw(s)
    };
}

