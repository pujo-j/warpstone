# Warpstone

Simple encrypted connection over websocket

Assumes an already secure TLS channel, do not trust the included crypto without a layer of conventional TLS over it !

Inside the already classically secure websocket over TLS channel,
 creates a new message channel secured by XSalsa20+Poly1305 with the shared key being the output of a combination of post-quantum SIKE and a pre shared key.
 
# Security

No formal proof whatsoever, use at your own risk, 
at worst it adds nothing to the existing TLS channel, 
but should not diminish its security, at best it's quantum proof 
even if the shared secret is leaked, at worst it will still protect against most non-state actors even in a post-quantum world.
 
