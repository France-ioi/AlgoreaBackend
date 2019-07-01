package tokentest

import "github.com/SermoDigital/jose/crypto"

// AlgoreaPlatformPublicKey represents a sample public key used to decode
// AlgoreaPlatform tokens in tests
var AlgoreaPlatformPublicKey = []byte(
	"-----BEGIN PUBLIC KEY-----\n" +
		"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDfsh3Rj/IAQ75LB7c8riFTYrgS\n" +
		"0FCDwZhYIPYgmqVWGPK7JX5KnrcTYqxr0e6nqD5e4anMIVyMUn7g+W9ULLa5QrFr\n" +
		"aJw7il+r1XPyadsPGe2C+YVqbSv33TRxTL03mzvlsLL+JlNvM7j0iJ/KGclLPHUz\n" +
		"fiE7YZDILwmultaYFQIDAQAB\n" +
		"-----END PUBLIC KEY-----")

// AlgoreaPlatformPublicKeyParsed represents a parsed sample public key used to decode
// AlgoreaPlatform tokens in tests
var AlgoreaPlatformPublicKeyParsed, _ = crypto.ParseRSAPublicKeyFromPEM(AlgoreaPlatformPublicKey)

// AlgoreaPlatformPrivateKey represents a sample private key used to encode
// AlgoreaPlatform tokens in tests
var AlgoreaPlatformPrivateKey = []byte(
	"-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIICXAIBAAKBgQDfsh3Rj/IAQ75LB7c8riFTYrgS0FCDwZhYIPYgmqVWGPK7JX5K\n" +
		"nrcTYqxr0e6nqD5e4anMIVyMUn7g+W9ULLa5QrFraJw7il+r1XPyadsPGe2C+YVq\n" +
		"bSv33TRxTL03mzvlsLL+JlNvM7j0iJ/KGclLPHUzfiE7YZDILwmultaYFQIDAQAB\n" +
		"AoGALEiomonykJbYnyXh4oNeWZGbey3+Inc634d28jFrNcYul1nuzHrrJ01LcPTY\n" +
		"WBx4bHQkFyMrnSPftk3q+jD34wpCEiBMFJmZk/Exj8ypRvN9K4+oJtMjvx3tcuyB\n" +
		"fnFRvf1J2sTL7F499xv+/UHAIGfyIvyYHLg/SV+aBaHDJmkCQQD3VqeDRTiMul5p\n" +
		"hDc4RbNLgWS3u1KT2U615OcTJZsFVzHuL6LhkxKLsc+rUWNurY0vOkwz4Bra2CpZ\n" +
		"klb/pVFvAkEA54eCYQ3UHUq+HUGFAX7fPokunjf9V+khU5PfvkzFI1O6DbvT5VCe\n" +
		"H4RVzM787lOy17TyIMvGqSIcLbf1hyekuwJAbUT6IlM9ZWaceS8xGgoo6K2Uals2\n" +
		"Yxz42gDzWREfCF/6Lgkbg15vLgny/fOp4uaHXhr6OVzDYHVpWEL/bleBvwJADwAS\n" +
		"jGMu+O7cvlx+V4h2wkB1Cr8p5MYv6JBOELA8nXtRNI6UveipNfWG8Yv/ixlVHvCU\n" +
		"N1e8eTzCgpvGhokk/QJBAJv1h/9jNOB9H9GIf3sB0cRLzH6po6aQX1gEYRZP6hIw\n" +
		"KGHLOGPIBt1FHY5Z0WtQ4vaFtwOEPj5BCPLGP9cvLIs=\n" +
		"-----END RSA PRIVATE KEY-----")

// AlgoreaPlatformPrivateKeyParsed represents a parsed sample private key used to encode
// AlgoreaPlatform tokens in tests
var AlgoreaPlatformPrivateKeyParsed, _ = crypto.ParseRSAPrivateKeyFromPEM(AlgoreaPlatformPrivateKey)

// TaskPlatformPublicKey represents a sample public key used to decode
// TaskPlatform tokens in tests
var TaskPlatformPublicKey = []byte(
	"-----BEGIN PUBLIC KEY-----\n" +
		"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDd04UmIYbKEyiu8bxbAIwM/4qU\n" +
		"RLHvVZHLExomx7snKjhy5Pft06Df+DzoRsTOA+zw8mFawwENFczjZ8b9vxS2Fecs\n" +
		"I7sNnrne8MJhNTjAHHX18C+qY+b+fA4enGTSGcZmxc1BxmFlpir1eck6/RSHUdj5\n" +
		"m1/97X7e9It4imr11wIDAQAB\n" +
		"-----END PUBLIC KEY-----")

// TaskPlatformPublicKeyParsed represents a parsed sample public key used to decode
// TaskPlatform tokens in tests
var TaskPlatformPublicKeyParsed, _ = crypto.ParseRSAPublicKeyFromPEM(TaskPlatformPublicKey)

// TaskPlatformPrivateKey represents a sample private key used to encode
// TaskPlatform tokens in tests
var TaskPlatformPrivateKey = []byte(
	"-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIICXQIBAAKBgQDd04UmIYbKEyiu8bxbAIwM/4qURLHvVZHLExomx7snKjhy5Pft\n" +
		"06Df+DzoRsTOA+zw8mFawwENFczjZ8b9vxS2FecsI7sNnrne8MJhNTjAHHX18C+q\n" +
		"Y+b+fA4enGTSGcZmxc1BxmFlpir1eck6/RSHUdj5m1/97X7e9It4imr11wIDAQAB\n" +
		"AoGAS/UG7dSjHATNdIIwFhUs37KNGFIhf6uUXa4v0UGpMzMA207OGrDRsH+HE55P\n" +
		"+59affLxJSwK6xkg7Gl3uSG6DDBNDrHENzv/+6XrBMA9Qv248cM8wztBFZRkcAzD\n" +
		"ZpKD9ptooMC3o4GOFjgdpAKcHGL3ujO8Ht5SuKF7iDjWHckCQQD2z1tyBcGvYp4k\n" +
		"f03/EZ/n9KFypYitId1nmHfx+dqpk42652lf/A40grjYPAyHxMNTgquT6tb5us6J\n" +
		"DH99114VAkEA5hYDSg8WKBd0M7vLhc0a5pR6X43eOdFn4cw95SmNEAC7gPnl792Q\n" +
		"j2D8qnN6f8J4UGFEAgYdPTxwSa6GAn/rOwJBAIQ7+umvbeNy+fnh/z7/CWa0qd+M\n" +
		"Ext3vnEnvnP2AxLCDLisDcgwesfllfW8zpXbdS+EHjuFIiLw1IGXIaOhxTUCQQC2\n" +
		"Z/UrnVI/bnidGuB6ruQIsOVjI6FtzOnCRJ09M/e1HB+KXJNB2jFkucsVhn8zEgU4\n" +
		"FCRKRnafuW57u3RaPvdJAkB/TqRhwsu8YEeynE5xATJHZosMwLNbB6FXzRrr6IZQ\n" +
		"TO0NDpRWmPbm+NVnfF+EPko+TyWlnsBi9aREWgn96Pgi\n" +
		"-----END RSA PRIVATE KEY-----")

// TaskPlatformPrivateKeyParsed represents a parsed sample private key used to encode
// TaskPlatform tokens in tests
var TaskPlatformPrivateKeyParsed, _ = crypto.ParseRSAPrivateKeyFromPEM(TaskPlatformPrivateKey)
