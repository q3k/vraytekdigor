use serde::{Deserialize};
use sha2::Digest;

/// 'Seeds' seem to be the main configuration system employed by DrayOS5. All data is stored in a
/// JSON file usually available as /etc/seeds/config_data.json, which is also committed to flash
/// memory by the firmware whenever the config is changed.
///
/// Apart from configuration data, the seeds directory also contains the definition of the
/// _configurability_ of the configuration, ie. the forms that are rendered to the user (and also
/// the settings that are visible in draysh). This module does not (yet) deal with that part of the
/// seeds functionality.
///
/// The meaning of the name 'seeds' is unknown, but DrayOS5 also has another subsystem named
/// 'feeds', which might mean they either like rhymes, are Simpsons fans, or are world-tier memers.
/// We will probably never know.

#[derive(Deserialize)]
pub struct Seeds {
    #[serde(rename = "0ADM_PASSWORD")]
    admin_passwords: Vec<AdminPassword>,
}

impl Seeds {
    /// parse data read from config_data.json into Seeds.
    pub fn parse(data: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(data)
    }

    /// Get an admin user by name. Currently there only seems to be one admin user, named admin.
    pub fn get_admin<'a, 'b>(&'a self, name: &'b str) -> Option<&'a AdminPassword> {
        for admin in &self.admin_passwords {
            if admin.name == name {
                return Some(admin);
            }
        }
        None
    }
}

/// Structure describing administrative users, store under 0ADM_PASSWORD in seeds.
#[derive(Deserialize)]
pub struct AdminPassword {
    #[serde(rename = "Name")]
    name: String,
    /// Password, hashed using... things. See the check_password function for more details.
    #[serde(rename = "Password")]
    password: String,
    /// SSH key. This field is not usually present in seeds data, but the CFW adds it (see the
    /// script in default.nix). It's thus marked as optional, so that libalien can work even
    /// without this modification.
    #[serde(rename = "SSHKey")]
    pub ssh_key: Option<String>,
}

impl AdminPassword {
    /// Returns true if the given password corresponds to the stored hash of this user.
    pub fn check_password(&self, password: &str) -> bool {
        // Split out wanted hash and wanted slalt from stored data.
        let parts: Vec<&str> = self.password.split("@").collect();
        // A third @-delimited field is always present (hash@salt@0), there might be more in the
        // future.
        if parts.len() < 2 {
            return false;
        }
        let hash = parts[0].clone();
        let salt = parts[1].clone();

        // First step: SHA-512 and hex-encode the given plaintext password.
        let mut hasher = sha2::Sha512::new();
        hasher.update(password.as_bytes());
        let sha512_hex = hex::encode(hasher.finalize());

        // Second step: PKBDF2, but funny. Instead of just pbkdf2'ing the plaintext password, or
        // even the sha512 hash above, we instead pkbdf2 the password and the requested salt.
        // Weird.
        let pbuf = format!("{}:{}", sha512_hex, salt);
        let mut res = [0u8; 64];
        pbkdf2::pbkdf2::<hmac::Hmac<sha2::Sha256>>(pbuf.as_bytes(), salt.as_bytes(), 1000, &mut res);

        // Encode the calculated hash to hex, instead of hex-decoding the stored hash. Less things
        // that can fail this way.
        let res_hex = hex::encode(res);

        // And compare the pbkdf2 hash with requested hash.
        return constant_time_eq::constant_time_eq(res_hex.as_bytes(), hash.as_bytes());
    }
}

#[cfg(test)]
mod test {
    use super::Seeds;

    #[test]
    fn test_seed_auth() {
        let data = r#"
            {
                "version": "1.0.0",
                "0OPERATION_MODE": [
                    {
                        "Name": "Setting",
                        "Mode": "Modem Mode"
                    }
                ],
                "0ADM_PASSWORD": [
                    {
                        "Name": "admin",
                        "Password": "5eb682fd41ede0fe47eba27267c0d653fa67efea1850b4f8a05e0cd7665383149e9ffeebc0b03e1449b7a7f510225efafd058d6341c911d6b242143128bdfa63@5c1759824f559e9258cd7801e3e350db@0"
                    }
                ]
            }
            "#;
        let s = Seeds::parse(data).expect("failed to parse seeds");
        let a = s.get_admin("admin").expect("failed to find admin user");
        assert!(a.check_password("admin"), "Password did not match");
    }
}
